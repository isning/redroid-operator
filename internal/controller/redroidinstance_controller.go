package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

const (
	instanceFinalizer = "redroid.io/instance-finalizer"
	defaultADBPort    = int32(5555)
)

// RedroidInstanceReconciler reconciles a RedroidInstance object.
// It ensures a Pod exists and is running when neither spec.suspend nor
// status.suspended is set, and deletes the Pod when either is true.
//
// Temporary suspend (status.suspended) is designed for programmatic use
// (e.g., RedroidTask with suspendInstance=true, or manual maintenance) without
// touching spec.suspend, so GitOps tools like Flux do not overwrite the state.
//
// +kubebuilder:rbac:groups=redroid.io,resources=redroidinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redroid.io,resources=redroidinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redroid.io,resources=redroidinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/portforward,verbs=create
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
type RedroidInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	// Auto-clear expired Suspended before making scheduling decisions.
	if ts := instance.Status.Suspended; ts != nil && ts.Until != nil && ts.Until.Time.Before(time.Now()) {
		patch := client.MergeFrom(instance.DeepCopy())
		instance.Status.Suspended = nil
		if err := r.Status().Patch(ctx, instance, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("clear expired suspended: %w", err)
		}
		logger.Info("cleared expired suspended", "instance", instance.Name)
	}

	if isEffectivelySuspended(instance) {
		// Desired state: no Pod running.
		if err := r.deleteInstancePod(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return r.updateStatus(ctx, instance, redroidv1alpha1.RedroidInstanceStopped, podName, "")
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
			return ctrl.Result{}, fmt.Errorf("create pod: %w", err)
		}
		return r.updateStatus(ctx, instance, redroidv1alpha1.RedroidInstancePending, podName, "")
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

	return r.updateStatus(ctx, instance, phase, podName, adbAddr)
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
	return client.IgnoreNotFound(r.Delete(ctx, pod))
}

func (r *RedroidInstanceReconciler) updateStatus(
	ctx context.Context,
	instance *redroidv1alpha1.RedroidInstance,
	phase redroidv1alpha1.RedroidInstancePhase,
	podName, adbAddr string,
) (ctrl.Result, error) {
	patch := client.MergeFrom(instance.DeepCopy())
	instance.Status.ObservedGeneration = instance.Generation
	instance.Status.Phase = phase
	instance.Status.PodName = podName
	instance.Status.ADBAddress = adbAddr

	// Sync Ready condition.
	readyStatus := metav1.ConditionFalse
	readyReason := string(phase)
	if phase == redroidv1alpha1.RedroidInstanceRunning && adbAddr != "" {
		readyStatus = metav1.ConditionTrue
		readyReason = "Running"
	}
	setCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidInstanceConditionReady),
		Status:             readyStatus,
		ObservedGeneration: instance.Generation,
		Reason:             readyReason,
		Message:            fmt.Sprintf("Phase=%s", phase),
	})

	// Sync Scheduled condition.
	scheduledStatus := metav1.ConditionFalse
	scheduledReason := "NoPod"
	if podName != "" && phase != redroidv1alpha1.RedroidInstanceStopped {
		scheduledStatus = metav1.ConditionTrue
		scheduledReason = "PodCreated"
	}
	setCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidInstanceConditionScheduled),
		Status:             scheduledStatus,
		ObservedGeneration: instance.Generation,
		Reason:             scheduledReason,
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

// isEffectivelySuspended returns true when the instance Pod should not be running,
// considering both the GitOps-managed spec.suspend and the programmatic
// status.suspended (set by tasks or manual operators).
// An expired suspended (Until in the past) is treated as cleared.
func isEffectivelySuspended(instance *redroidv1alpha1.RedroidInstance) bool {
	if instance.Spec.Suspend {
		return true
	}
	ts := instance.Status.Suspended
	if ts == nil {
		return false
	}
	// Expired? Treat as not suspended — the reconciler clears it on the same pass.
	if ts.Until != nil && ts.Until.Time.Before(time.Now()) {
		return false
	}
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

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instancePodName(instance),
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "redroid-operator",
				"redroid.io/instance":          instance.Name,
				"redroid.io/instance-index":    indexStr,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:    corev1.RestartPolicyNever,
			NodeSelector:     instance.Spec.NodeSelector,
			Tolerations:      instance.Spec.Tolerations,
			Affinity:         instance.Spec.Affinity,
			ImagePullSecrets: instance.Spec.ImagePullSecrets,
			Containers: []corev1.Container{
				{
					Name:            "redroid",
					Image:           instance.Spec.Image,
					ImagePullPolicy: pullPolicy,
					Args:            args,
					Resources:       instance.Spec.Resources,
					SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
					Ports: []corev1.ContainerPort{
						{ContainerPort: port, Protocol: corev1.ProtocolTCP},
					},
					Env: append([]corev1.EnvVar{
						{Name: "INSTANCE_INDEX", Value: indexStr},
					}, instance.Spec.ExtraEnv...),
					VolumeMounts: volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}
	return pod
}

func instancePodName(instance *redroidv1alpha1.RedroidInstance) string {
	return fmt.Sprintf("redroid-instance-%s", instance.Name)
}

func instanceServiceName(instance *redroidv1alpha1.RedroidInstance) string {
	return fmt.Sprintf("redroid-instance-%s", instance.Name)
}

// reconcileService ensures a stable Service exists for the instance.
// The Service selector targets the Pod label redroid.io/instance=<name>, which
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
				"redroid.io/instance":          instance.Name,
			},
			Annotations: extraAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Type: svcType,
			Selector: map[string]string{
				"redroid.io/instance": instance.Name,
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
