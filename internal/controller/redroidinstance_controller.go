package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

const (
	instanceFinalizer = "redroid.isning.moe/instance-finalizer"
	defaultADBPort    = int32(5555)
)

// RedroidInstanceReconciler reconciles a RedroidInstance object.
// It ensures a Pod exists and is running when the instance should be running,
// and deletes the Pod when it should be stopped.
//
// Phase decision (highest priority first):
//  1. status.woken non-nil and not expired → Running  (on-demand wake, overrides spec.suspend)
//  2. spec.suspend == true                 → Stopped  (permanent user intent)
//  3. status.suspended non-nil, not expired → Stopped (programmatic / maintenance override)
//  4. default                              → Running
//
// Temporary overrides (status.suspended, status.woken) live in status so GitOps tools
// like Flux do not reconcile them back, avoiding config drift.
//
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/portforward,verbs=create
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
type RedroidInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile reconciles a RedroidInstance object.
func (r *RedroidInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &redroidv1alpha1.RedroidInstance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion: clean up Pod before removing finalizer.
	if !instance.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(instance, instanceFinalizer) {
			if err := r.deleteInstancePod(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(instance, instanceFinalizer)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Ensure finalizer.
	if !controllerutil.ContainsFinalizer(instance, instanceFinalizer) {
		controllerutil.AddFinalizer(instance, instanceFinalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	podName := instancePodName(instance)

	// Always ensure the Service exists (stable DNS / port-forward target).
	if err := r.reconcileService(ctx, instance); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile service: %w", err)
	}

	// Auto-clear expired Woken.
	if w := instance.Status.Woken; w != nil && w.Until != nil && w.Until.Time.Before(time.Now()) {
		patch := client.MergeFrom(instance.DeepCopy())
		instance.Status.Woken = nil
		if err := r.Status().Patch(ctx, instance, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("clear expired woken: %w", err)
		}
		logger.Info("cleared expired woken", "instance", instance.Name)
		r.Recorder.Event(instance, corev1.EventTypeNormal, "WokenExpired", "Cleared expired programmatic wake override")
	}

	// Auto-clear expired Suspended before making scheduling decisions.
	if ts := instance.Status.Suspended; ts != nil && ts.Until != nil && ts.Until.Time.Before(time.Now()) {
		patch := client.MergeFrom(instance.DeepCopy())
		instance.Status.Suspended = nil
		if err := r.Status().Patch(ctx, instance, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("clear expired suspended: %w", err)
		}
		logger.Info("cleared expired suspended", "instance", instance.Name)
		r.Recorder.Event(instance, corev1.EventTypeNormal, "SuspendedExpired", "Cleared expired programmatic suspend override")
	}

	if !desiredRunning(instance) {
		// Desired state: no Pod running.
		if err := r.deleteInstancePod(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		// The event for PodDeleted is handled inside deleteInstancePod if a
		// Pod actually existed to be deleted.
		return r.updateStatus(ctx, instance, redroidv1alpha1.RedroidInstanceStopped, podName, "", nil)
	}

	// Desired state: Pod running.
	existingPod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: instance.Namespace}, existingPod)
	if apierrors.IsNotFound(err) {
		logger.Info("creating Pod for RedroidInstance", "pod", podName)
		pod := r.buildInstancePod(instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, pod); err != nil {
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "FailedCreatePod", "Failed to create Pod %s: %v", podName, err)
			return ctrl.Result{}, fmt.Errorf("create pod: %w", err)
		}
		r.Recorder.Eventf(instance, corev1.EventTypeNormal, "CreatedPod", "Created Pod %s", podName)
		return r.updateStatus(ctx, instance, redroidv1alpha1.RedroidInstancePending, podName, "", nil)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// Pod exists — sync status.
	phase := redroidv1alpha1.RedroidInstancePending
	adbAddr := ""
	port := instanceADBPort(instance)
	switch existingPod.Status.Phase {
	case corev1.PodRunning:
		phase = redroidv1alpha1.RedroidInstanceRunning
		if existingPod.Status.PodIP != "" {
			adbAddr = fmt.Sprintf("%s:%d", existingPod.Status.PodIP, port)
		}
	case corev1.PodFailed:
		phase = redroidv1alpha1.RedroidInstanceFailed
	case corev1.PodSucceeded:
		// Pod exited normally — treat as stopped.
		phase = redroidv1alpha1.RedroidInstanceStopped
	}

	return r.updateStatus(ctx, instance, phase, podName, adbAddr, existingPod)
}

func (r *RedroidInstanceReconciler) deleteInstancePod(ctx context.Context, instance *redroidv1alpha1.RedroidInstance) error {
	pod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      instancePodName(instance),
		Namespace: instance.Namespace,
	}, pod)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := r.Delete(ctx, pod); client.IgnoreNotFound(err) != nil {
		return err
	}
	r.Recorder.Eventf(instance, corev1.EventTypeNormal, "DeletedPod", "Deleted Pod %s", pod.Name)
	return nil
}

// podFailureDetails extracts a human-readable failure summary from a Pod's
// container statuses. It returns an empty string when no failure information
// is available (e.g. the Pod has not yet terminated).
func podFailureDetails(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}
	var parts []string
	for _, cs := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		term := cs.State.Terminated
		if term == nil || term.ExitCode == 0 {
			continue
		}
		detail := fmt.Sprintf("container %q exited with code %d", cs.Name, term.ExitCode)
		if term.Reason != "" {
			detail += fmt.Sprintf(" (reason: %s)", term.Reason)
		}
		if term.Message != "" {
			detail += fmt.Sprintf(": %s", term.Message)
		}
		parts = append(parts, detail)
	}
	if len(parts) == 0 {
		if pod.Status.Message != "" {
			return pod.Status.Message
		}
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += "; " + p
	}
	return result
}

func (r *RedroidInstanceReconciler) updateStatus(
	ctx context.Context,
	instance *redroidv1alpha1.RedroidInstance,
	phase redroidv1alpha1.RedroidInstancePhase,
	podName, adbAddr string,
	pod *corev1.Pod,
) (ctrl.Result, error) {
	oldPhase := instance.Status.Phase

	patch := client.MergeFrom(instance.DeepCopy())
	instance.Status.ObservedGeneration = instance.Generation
	instance.Status.Phase = phase
	instance.Status.PodName = podName
	instance.Status.ADBAddress = adbAddr

	// Sync Ready condition.
	readyStatus := metav1.ConditionFalse
	var readyReason, readyMsg string
	switch phase {
	case redroidv1alpha1.RedroidInstanceRunning:
		if adbAddr != "" {
			readyStatus = metav1.ConditionTrue
			readyReason = "Running"
			readyMsg = fmt.Sprintf("Pod %q is running; ADB available at %s.", podName, adbAddr)
			if oldPhase != phase {
				r.Recorder.Eventf(instance, corev1.EventTypeNormal, "InstanceReady", "ADB port is available at %s", adbAddr)
			}
		} else {
			readyReason = "PodRunningNoADB"
			readyMsg = fmt.Sprintf("Pod %q is running, but its IP address is not yet available.", podName)
		}
	case redroidv1alpha1.RedroidInstanceStopped:
		readyReason = "Stopped"
		if instance.Spec.Suspend {
			readyMsg = "Instance is stopped because spec.suspend is true."
		} else if instance.Status.Suspended != nil {
			actor := instance.Status.Suspended.Actor
			if actor == "" {
				actor = "unknown"
			}
			readyMsg = fmt.Sprintf("Instance is temporarily suspended by %q (status.suspended).", actor)
		} else {
			readyMsg = "Instance is stopped."
		}
	case redroidv1alpha1.RedroidInstanceFailed:
		readyReason = "PodFailed"
		details := podFailureDetails(pod)
		if details != "" {
			readyMsg = fmt.Sprintf("Pod %q failed: %s", podName, details)
		} else {
			readyMsg = fmt.Sprintf("Pod %q has failed. Inspect pod events and logs for details.", podName)
		}
		if oldPhase != phase {
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "InstanceFailed", readyMsg)
		}
	case redroidv1alpha1.RedroidInstancePending:
		readyReason = "Pending"
		if podName != "" {
			readyMsg = fmt.Sprintf("Pod %q is pending scheduling or startup.", podName)
		} else {
			readyMsg = "Pod is being created."
		}
	default:
		readyReason = string(phase)
		readyMsg = fmt.Sprintf("Instance phase is %s.", phase)
	}
	setCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidInstanceConditionReady),
		Status:             readyStatus,
		ObservedGeneration: instance.Generation,
		Reason:             readyReason,
		Message:            readyMsg,
	})

	// Sync Scheduled condition.
	scheduledStatus := metav1.ConditionFalse
	var scheduledReason, scheduledMsg string
	switch {
	case phase == redroidv1alpha1.RedroidInstanceStopped:
		scheduledReason = "Stopped"
		scheduledMsg = "Instance is stopped; no Pod is scheduled."
	case podName != "" && phase == redroidv1alpha1.RedroidInstanceFailed:
		scheduledStatus = metav1.ConditionTrue
		scheduledReason = "PodCreated"
		scheduledMsg = fmt.Sprintf("Pod %q was created but has since failed.", podName)
	case podName != "":
		scheduledStatus = metav1.ConditionTrue
		scheduledReason = "PodCreated"
		scheduledMsg = fmt.Sprintf("Pod %q has been created and scheduled.", podName)
	default:
		scheduledReason = "NoPod"
		scheduledMsg = "No Pod has been created yet."
	}
	setCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidInstanceConditionScheduled),
		Status:             scheduledStatus,
		ObservedGeneration: instance.Generation,
		Reason:             scheduledReason,
		Message:            scheduledMsg,
	})

	return ctrl.Result{}, r.Status().Patch(ctx, instance, patch)
}

// setCondition upserts a condition into the slice by Type.
func setCondition(conditions *[]metav1.Condition, cond metav1.Condition) {
	now := metav1.Now()
	for i, c := range *conditions {
		if c.Type == cond.Type {
			if c.Status != cond.Status {
				cond.LastTransitionTime = now
			} else {
				cond.LastTransitionTime = c.LastTransitionTime
			}
			(*conditions)[i] = cond
			return
		}
	}
	cond.LastTransitionTime = now
	*conditions = append(*conditions, cond)
}

// instanceADBPort returns the configured ADB port or the default 5555.
func instanceADBPort(instance *redroidv1alpha1.RedroidInstance) int32 {
	if instance.Spec.ADBPort != nil {
		return *instance.Spec.ADBPort
	}
	return defaultADBPort
}

// desiredRunning returns true when the instance Pod should be running, applying
// the four-level priority rule documented on RedroidInstanceReconciler.
func desiredRunning(instance *redroidv1alpha1.RedroidInstance) bool {
	// 1. Woken override — beats spec.suspend.
	if w := instance.Status.Woken; w != nil {
		if w.Until == nil || !w.Until.Time.Before(time.Now()) {
			return true
		}
		// expired — fall through
	}
	// 2. Permanent user intent.
	if instance.Spec.Suspend {
		return false
	}
	// 3. Programmatic / maintenance override.
	if s := instance.Status.Suspended; s != nil {
		if s.Until == nil || !s.Until.Time.Before(time.Now()) {
			return false
		}
		// expired — fall through
	}
	// 4. Default: run.
	return true
}

// buildInstancePod constructs the Pod manifest for a RedroidInstance.
func (r *RedroidInstanceReconciler) buildInstancePod(instance *redroidv1alpha1.RedroidInstance) *corev1.Pod {
	indexStr := fmt.Sprintf("%d", instance.Spec.Index)
	port := instanceADBPort(instance)

	args := buildRedroidArgs(instance.Spec)

	pullPolicy := instance.Spec.ImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}

	privileged := true
	hostPathType := corev1.HostPathDirectoryOrCreate

	// Volume mounts differ between normal (overlayfs) and base-init mode.
	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume

	if instance.Spec.BaseMode {
		// Base mode: mount sharedDataPVC directly as /data for first-boot initialisation.
		volumeMounts = []corev1.VolumeMount{
			{Name: "data-base", MountPath: "/data"},
			{Name: "dev-dri", MountPath: "/dev/dri"},
		}
		volumes = []corev1.Volume{
			{
				Name: "data-base",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: instance.Spec.SharedDataPVC,
					},
				},
			},
			{
				Name: "dev-dri",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/dri",
						Type: &hostPathType,
					},
				},
			},
		}
	} else {
		// Normal mode: overlayfs with shared lower layer and per-instance upper layer.
		volumeMounts = []corev1.VolumeMount{
			{Name: "data-base", MountPath: "/data-base"},
			{Name: "data-diff", MountPath: "/data-diff/" + indexStr, SubPath: indexStr},
			{Name: "dev-dri", MountPath: "/dev/dri"},
		}
		volumes = []corev1.Volume{
			{
				Name: "data-base",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: instance.Spec.SharedDataPVC,
					},
				},
			},
			{
				Name: "data-diff",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: instance.Spec.DiffDataPVC,
					},
				},
			},
			{
				Name: "dev-dri",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/dri",
						Type: &hostPathType,
					},
				},
			},
		}
	}

	// When kmsg redirect is enabled, add a shared emptyDir for PID/ready handshake.
	if !instance.Spec.DisableKmsgRedirect {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      kmsgSyncVolume,
			MountPath: kmsgSyncMount,
		})
		volumes = append(volumes, corev1.Volume{
			Name: kmsgSyncVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	initContainers, containers := buildInitAndMainContainers(instance, args, pullPolicy, port, indexStr, privileged, volumeMounts)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instancePodName(instance),
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":      "redroid-operator",
				"redroid.isning.moe/instance":       instance.Name,
				"redroid.isning.moe/instance-index": indexStr,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:    corev1.RestartPolicyNever,
			NodeSelector:     instance.Spec.NodeSelector,
			Tolerations:      instance.Spec.Tolerations,
			Affinity:         instance.Spec.Affinity,
			ImagePullSecrets: instance.Spec.ImagePullSecrets,
			InitContainers:   initContainers,
			Containers:       containers,
			Volumes:          volumes,
		},
	}
	return pod
}

// kmsgSyncVolume is the emptyDir volume name used to share the socat binary
// between the init container and the main container.
const kmsgSyncVolume = "kmsg-tools"

// kmsgSyncMount is the path at which kmsgSyncVolume is mounted in both containers.
const kmsgSyncMount = "/kmsg-tools"

// kmsgMainWrapper is injected as the main container's shell command.
// It uses the locally copied socat binary to create a PTY, bind-mounts its
// slave end over /dev/kmsg, and forwards the PTY master output to stdout.
// This ensures Android kernel logs are accessible via `kubectl logs <pod>`
// and prevents host dmesg pollution, all without requiring a sidecar.
const kmsgMainWrapper = `
/kmsg-tools/socat PTY,link=/tmp/kmsg-pty,mode=0622,rawer - &
SOCAT_PID=$!
/kmsg-tools/busybox sleep 0.3
/kmsg-tools/busybox mount --bind /tmp/kmsg-pty /dev/kmsg
exec /init "$@"
`

// buildInitAndMainContainers constructs the container list for the Pod.
// When kmsg redirect is enabled (the default) it returns two containers:
//   - An init container that copies a musl-static socat binary into an emptyDir.
//   - The main "redroid" container which runs a wrapper script that uses the
//     injected socat to redirect /dev/kmsg before exec'ing /init.
func buildInitAndMainContainers(
	instance *redroidv1alpha1.RedroidInstance,
	args []string,
	pullPolicy corev1.PullPolicy,
	port int32,
	indexStr string,
	privileged bool,
	volumeMounts []corev1.VolumeMount,
) (initContainers []corev1.Container, containers []corev1.Container) {
	// Main container command/args:
	//   - disabled: use image ENTRYPOINT (/init) with bare androidboot args
	//   - enabled:  /kmsg-tools/busybox sh wrapper sets up socat, then exec /init "$@"
	mainCmd := []string(nil)
	mainArgs := args
	if !instance.Spec.DisableKmsgRedirect {
		mainCmd = []string{"/kmsg-tools/busybox", "sh"}
		mainArgs = append([]string{"-c", kmsgMainWrapper, "--"}, args...)
	}

	main := corev1.Container{
		Name:            "redroid",
		Image:           instance.Spec.Image,
		ImagePullPolicy: pullPolicy,
		Command:         mainCmd,
		Args:            mainArgs,
		Resources:       instance.Spec.Resources,
		SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
		Ports: []corev1.ContainerPort{
			{ContainerPort: port, Protocol: corev1.ProtocolTCP},
		},
		Env: append([]corev1.EnvVar{
			{Name: "INSTANCE_INDEX", Value: indexStr},
		}, instance.Spec.ExtraEnv...),
		VolumeMounts: volumeMounts,
	}

	if instance.Spec.DisableKmsgRedirect {
		return nil, []corev1.Container{main}
	}

	// Init container copies socat to the shared emptyDir.
	toolsImg := os.Getenv("RELATED_IMAGE_KMSG_TOOLS")
	if toolsImg == "" {
		toolsImg = "ghcr.io/isning/redroid-operator/kmsg-tools:latest"
	}
	if instance.Spec.KmsgToolsImage != "" {
		toolsImg = instance.Spec.KmsgToolsImage
	}

	pullPolicyInit := corev1.PullIfNotPresent
	lastColon := strings.LastIndex(toolsImg, ":")
	lastSlash := strings.LastIndex(toolsImg, "/")
	if strings.HasSuffix(toolsImg, ":latest") || lastColon <= lastSlash {
		pullPolicyInit = corev1.PullAlways
	}

	initContainer := corev1.Container{
		Name:            "kmsg-tools",
		Image:           toolsImg,
		ImagePullPolicy: pullPolicyInit,
		Command:         []string{"/bin/sh", "-c", "cp /kmsg-bin/socat /kmsg-bin/busybox /kmsg-tools/ || exit 1; chmod +x /kmsg-tools/socat /kmsg-tools/busybox"},
		VolumeMounts: []corev1.VolumeMount{
			{Name: kmsgSyncVolume, MountPath: kmsgSyncMount},
		},
	}

	return []corev1.Container{initContainer}, []corev1.Container{main}
}

func instancePodName(instance *redroidv1alpha1.RedroidInstance) string {
	return fmt.Sprintf("redroid-instance-%s", instance.Name)
}

func instanceServiceName(instance *redroidv1alpha1.RedroidInstance) string {
	return fmt.Sprintf("redroid-instance-%s", instance.Name)
}

// reconcileService ensures a stable Service exists for the instance.
// The Service selector targets the Pod label redroid.isning.moe/instance=<name>, which
// means port-forward works via the Service rather than a specific Pod — the CLI
// does not need to know or store the Pod name.
func (r *RedroidInstanceReconciler) reconcileService(ctx context.Context, instance *redroidv1alpha1.RedroidInstance) error {
	port := instanceADBPort(instance)

	// Derive Service settings from spec.service (with safe defaults).
	svcType := corev1.ServiceTypeClusterIP
	var extraAnnotations map[string]string
	var nodePort int32
	if s := instance.Spec.Service; s != nil {
		if s.Type != "" {
			svcType = s.Type
		}
		extraAnnotations = s.Annotations
		if s.NodePort != nil {
			nodePort = *s.NodePort
		}
	}

	svcPort := corev1.ServicePort{Name: "adb", Port: port, Protocol: corev1.ProtocolTCP}
	if nodePort > 0 {
		svcPort.NodePort = nodePort
	}

	want := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceServiceName(instance),
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "redroid-operator",
				"redroid.isning.moe/instance":  instance.Name,
			},
			Annotations: extraAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Type: svcType,
			Selector: map[string]string{
				"redroid.isning.moe/instance": instance.Name,
			},
			Ports: []corev1.ServicePort{svcPort},
		},
	}

	existing := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: want.Name, Namespace: want.Namespace}, existing)
	if apierrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, want, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, want)
	}
	if err != nil {
		return err
	}
	// Update mutable fields if they drifted.
	patch := client.MergeFrom(existing.DeepCopy())
	existing.Annotations = want.Annotations
	existing.Spec.Type = want.Spec.Type
	existing.Spec.Selector = want.Spec.Selector
	existing.Spec.Ports = want.Spec.Ports
	return r.Patch(ctx, existing, patch)
}

// SetupWithManager registers the controller with the Manager.
func (r *RedroidInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redroidv1alpha1.RedroidInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
