package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
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
	taskFinalizer = "redroid.io/task-finalizer"

	// adbReadyScript is injected as the entrypoint for integration containers.
	// It waits for the redroid ADB port to accept connections before running the real command.
	adbReadyScript = `
set -e
echo "[redroid-operator] Waiting for ADB on ${ADB_ADDRESS}..."
timeout=120; elapsed=0
until adb connect "${ADB_ADDRESS}"; do
  if [ "$elapsed" -ge "$timeout" ]; then
    echo "[redroid-operator] Timed out waiting for ADB, exiting."
    exit 1
  fi
  sleep 2; elapsed=$((elapsed + 2))
done
adb wait-for-device
echo "[redroid-operator] ADB ready. Waiting 30s for Android system..."
sleep 30
`
)

// RedroidTaskReconciler reconciles a RedroidTask object.
//
// For one-shot tasks (spec.schedule == ""), it creates one Job per InstanceRef.
// For scheduled tasks (spec.schedule != ""), it creates one CronJob per InstanceRef.
// All child resources are owned by the task and garbage-collected on deletion.
//
// +kubebuilder:rbac:groups=redroid.io,resources=redroidtasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redroid.io,resources=redroidtasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redroid.io,resources=redroidtasks/finalizers,verbs=update
// +kubebuilder:rbac:groups=redroid.io,resources=redroidinstances,verbs=get;list;watch
// +kubebuilder:rbac:groups=redroid.io,resources=redroidinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
type RedroidTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles a RedroidTask object.
func (r *RedroidTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	task := &redroidv1alpha1.RedroidTask{}
	if err := r.Get(ctx, req.NamespacedName, task); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !task.DeletionTimestamp.IsZero() {
		controllerutil.RemoveFinalizer(task, taskFinalizer)
		return ctrl.Result{}, r.Update(ctx, task)
	}

	if !controllerutil.ContainsFinalizer(task, taskFinalizer) {
		controllerutil.AddFinalizer(task, taskFinalizer)
		if err := r.Update(ctx, task); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Resolve each InstanceRef to a RedroidInstance.
	instances, err := r.resolveInstances(ctx, task)
	if err != nil {
		return ctrl.Result{}, err
	}

	if task.Spec.Schedule == "" {
		// One-shot: create one Job per instance.
		return r.reconcileJobs(ctx, task, instances, logger)
	}
	// Scheduled: create one CronJob per instance.
	return r.reconcileCronJobs(ctx, task, instances, logger)
}

// resolveInstances fetches all RedroidInstance resources referenced by the task.
func (r *RedroidTaskReconciler) resolveInstances(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
) ([]*redroidv1alpha1.RedroidInstance, error) {
	result := make([]*redroidv1alpha1.RedroidInstance, 0, len(task.Spec.Instances))
	for _, ref := range task.Spec.Instances {
		inst := &redroidv1alpha1.RedroidInstance{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      ref.Name,
			Namespace: task.Namespace,
		}, inst); err != nil {
			return nil, fmt.Errorf("resolve instance %q: %w", ref.Name, err)
		}
		result = append(result, inst)
	}
	return result, nil
}

// reconcileJobs ensures one Job per instance exists for a one-shot task.
// When task.Spec.SuspendInstance is true the controller first sets
// status.suspended on each instance (so the instance controller stops
// the Pod), waits for the instance to reach phase Stopped, then creates the
// Job.  After the Job finishes the temporary suspend is cleared.
//
// When task.Spec.WakeInstance is true the controller sets status.woken on each
// instance (so the instance controller starts the Pod even when spec.suspend is
// true), waits for phase Running, then creates the Job.  After the Job finishes
// status.woken is cleared, allowing spec.suspend to take effect again.
func (r *RedroidTaskReconciler) reconcileJobs(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
	instances []*redroidv1alpha1.RedroidInstance,
	logger interface{ Info(string, ...interface{}) },
) (ctrl.Result, error) {
	activeJobs := []string{}
	requeueAfter := time.Duration(0)

	for i, inst := range instances {
		jobName := fmt.Sprintf("%s-%s", task.Name, inst.Name)
		job := &batchv1.Job{}
		err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: task.Namespace}, job)

		if apierrors.IsNotFound(err) {
			if task.Spec.SuspendInstance {
				// Step 1: set suspended if not yet set.
				if inst.Status.Suspended == nil {
					if setErr := r.setInstanceSuspended(ctx, inst,
						"task/"+task.Name,
						"reserved for one-shot task "+task.Name,
					); setErr != nil {
						return ctrl.Result{}, setErr
					}
					logger.Info("set suspended on instance; waiting for Pod to stop",
						"instance", inst.Name)
					requeueAfter = 5 * time.Second
					continue
				}
				// Step 2: wait for instance Pod to stop.
				if inst.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
					logger.Info("waiting for instance to stop before creating Job",
						"instance", inst.Name, "phase", inst.Status.Phase)
					requeueAfter = 5 * time.Second
					continue
				}
			} else if task.Spec.WakeInstance {
				// Step 1: set woken if not yet set.
				if inst.Status.Woken == nil {
					if setErr := r.setInstanceWoken(ctx, inst,
						"task/"+task.Name,
						"on-demand wake for one-shot task "+task.Name,
					); setErr != nil {
						return ctrl.Result{}, setErr
					}
					logger.Info("set woken on instance; waiting for Pod to start",
						"instance", inst.Name)
					requeueAfter = 5 * time.Second
					continue
				}
				// Step 2: wait for instance Pod to start.
				if inst.Status.Phase != redroidv1alpha1.RedroidInstanceRunning {
					logger.Info("waiting for instance to start before creating Job",
						"instance", inst.Name, "phase", inst.Status.Phase)
					requeueAfter = 5 * time.Second
					continue
				}
			}
			logger.Info("creating Job for instance", "job", jobName, "instance", inst.Name)
			newJob := r.buildJob(task, inst, jobName, i)
			if err := controllerutil.SetControllerReference(task, newJob, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, newJob); err != nil {
				return ctrl.Result{}, fmt.Errorf("create job %s: %w", jobName, err)
			}
			activeJobs = append(activeJobs, jobName)
			continue
		}
		if err != nil {
			return ctrl.Result{}, err
		}

		if !isJobFinished(job) {
			activeJobs = append(activeJobs, jobName)
		} else {
			// Job finished — restore instance state.
			if task.Spec.SuspendInstance && inst.Status.Suspended != nil {
				if clrErr := r.clearInstanceSuspended(ctx, inst); clrErr != nil {
					return ctrl.Result{}, clrErr
				}
				logger.Info("cleared suspended after Job completion",
					"instance", inst.Name, "job", jobName)
			}
			if task.Spec.WakeInstance && inst.Status.Woken != nil {
				if clrErr := r.clearInstanceWoken(ctx, inst); clrErr != nil {
					return ctrl.Result{}, clrErr
				}
				logger.Info("cleared woken after Job completion",
					"instance", inst.Name, "job", jobName)
			}
		}
	}

	if requeueAfter > 0 {
		if _, patchErr := r.patchTaskStatus(ctx, task, activeJobs); patchErr != nil {
			return ctrl.Result{}, patchErr
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	return r.patchTaskStatus(ctx, task, activeJobs)
}

// setInstanceSuspended patches status.suspended on the instance.
func (r *RedroidTaskReconciler) setInstanceSuspended(
	ctx context.Context,
	inst *redroidv1alpha1.RedroidInstance,
	actor, reason string,
) error {
	patch := client.MergeFrom(inst.DeepCopy())
	inst.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
		Actor:  actor,
		Reason: reason,
	}
	return r.Status().Patch(ctx, inst, patch)
}

// clearInstanceSuspended removes status.suspended from the instance.
func (r *RedroidTaskReconciler) clearInstanceSuspended(
	ctx context.Context,
	inst *redroidv1alpha1.RedroidInstance,
) error {
	patch := client.MergeFrom(inst.DeepCopy())
	inst.Status.Suspended = nil
	return r.Status().Patch(ctx, inst, patch)
}

// setInstanceWoken patches status.woken on the instance, forcing it Running
// even when spec.suspend is true.
func (r *RedroidTaskReconciler) setInstanceWoken(
	ctx context.Context,
	inst *redroidv1alpha1.RedroidInstance,
	actor, reason string,
) error {
	patch := client.MergeFrom(inst.DeepCopy())
	inst.Status.Woken = &redroidv1alpha1.WokenStatus{
		Actor:  actor,
		Reason: reason,
	}
	return r.Status().Patch(ctx, inst, patch)
}

// clearInstanceWoken removes status.woken from the instance.
func (r *RedroidTaskReconciler) clearInstanceWoken(
	ctx context.Context,
	inst *redroidv1alpha1.RedroidInstance,
) error {
	patch := client.MergeFrom(inst.DeepCopy())
	inst.Status.Woken = nil
	return r.Status().Patch(ctx, inst, patch)
}

// reconcileCronJobs ensures one CronJob per instance exists for a scheduled task.
func (r *RedroidTaskReconciler) reconcileCronJobs(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
	instances []*redroidv1alpha1.RedroidInstance,
	logger interface{ Info(string, ...interface{}) },
) (ctrl.Result, error) {
	for i, inst := range instances {
		cronName := fmt.Sprintf("%s-%s", task.Name, inst.Name)
		existing := &batchv1.CronJob{}
		err := r.Get(ctx, types.NamespacedName{Name: cronName, Namespace: task.Namespace}, existing)

		desired := r.buildCronJob(task, inst, cronName, i)
		if err := controllerutil.SetControllerReference(task, desired, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if apierrors.IsNotFound(err) {
			logger.Info("creating CronJob for instance", "cronjob", cronName, "instance", inst.Name)
			if err := r.Create(ctx, desired); err != nil {
				return ctrl.Result{}, fmt.Errorf("create cronjob %s: %w", cronName, err)
			}
			continue
		}
		if err != nil {
			return ctrl.Result{}, err
		}

		// Update mutable fields from task spec.
		patch := client.MergeFrom(existing.DeepCopy())
		existing.Spec.Schedule = task.Spec.Schedule
		existing.Spec.Suspend = &task.Spec.Suspend
		if task.Spec.Timezone != "" {
			existing.Spec.TimeZone = &task.Spec.Timezone
		}
		if err := r.Patch(ctx, existing, patch); err != nil {
			return ctrl.Result{}, fmt.Errorf("patch cronjob %s: %w", cronName, err)
		}
	}

	return r.patchTaskStatus(ctx, task, nil)
}

func (r *RedroidTaskReconciler) patchTaskStatus(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
	activeJobs []string,
) (ctrl.Result, error) {
	patch := client.MergeFrom(task.DeepCopy())
	task.Status.ObservedGeneration = task.Generation
	task.Status.ActiveJobs = activeJobs
	return ctrl.Result{}, r.Status().Patch(ctx, task, patch)
}

// buildJob constructs a Job manifest that runs all integrations against a single Redroid instance.
func (r *RedroidTaskReconciler) buildJob(
	task *redroidv1alpha1.RedroidTask,
	inst *redroidv1alpha1.RedroidInstance,
	name string,
	_ int,
) *batchv1.Job {
	podSpec := r.buildTaskPodSpec(task, inst)

	backoffLimit := int32(0)
	if task.Spec.BackoffLimit != nil {
		backoffLimit = *task.Spec.BackoffLimit
	}

	jobSpec := batchv1.JobSpec{
		BackoffLimit: &backoffLimit,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: taskLabels(task, inst)},
			Spec:       podSpec,
		},
	}
	if task.Spec.ActiveDeadlineSeconds != nil {
		jobSpec.ActiveDeadlineSeconds = task.Spec.ActiveDeadlineSeconds
	}
	if task.Spec.TTLSecondsAfterFinished != nil {
		jobSpec.TTLSecondsAfterFinished = task.Spec.TTLSecondsAfterFinished
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: task.Namespace,
			Labels:    taskLabels(task, inst),
		},
		Spec: jobSpec,
	}
}

// buildCronJob constructs a CronJob that fires the per-instance pod on schedule.
func (r *RedroidTaskReconciler) buildCronJob(
	task *redroidv1alpha1.RedroidTask,
	inst *redroidv1alpha1.RedroidInstance,
	name string,
	_ int,
) *batchv1.CronJob {
	podSpec := r.buildTaskPodSpec(task, inst)

	backoffLimit := int32(0)
	if task.Spec.BackoffLimit != nil {
		backoffLimit = *task.Spec.BackoffLimit
	}

	successLimit := ptr(int32(3))
	failLimit := ptr(int32(3))
	if task.Spec.SuccessfulJobsHistoryLimit != nil {
		successLimit = task.Spec.SuccessfulJobsHistoryLimit
	}
	if task.Spec.FailedJobsHistoryLimit != nil {
		failLimit = task.Spec.FailedJobsHistoryLimit
	}

	suspend := task.Spec.Suspend
	cronSpec := batchv1.CronJobSpec{
		Schedule:                   task.Spec.Schedule,
		ConcurrencyPolicy:          batchv1.ForbidConcurrent,
		Suspend:                    &suspend,
		SuccessfulJobsHistoryLimit: successLimit,
		FailedJobsHistoryLimit:     failLimit,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: taskLabels(task, inst)},
			Spec: batchv1.JobSpec{
				BackoffLimit: &backoffLimit,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: taskLabels(task, inst)},
					Spec:       podSpec,
				},
			},
		},
	}
	if task.Spec.ActiveDeadlineSeconds != nil {
		cronSpec.JobTemplate.Spec.ActiveDeadlineSeconds = task.Spec.ActiveDeadlineSeconds
	}
	if task.Spec.StartingDeadlineSeconds != nil {
		cronSpec.StartingDeadlineSeconds = task.Spec.StartingDeadlineSeconds
	}
	if task.Spec.Timezone != "" {
		cronSpec.TimeZone = &task.Spec.Timezone
	}

	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: task.Namespace,
			Labels:    taskLabels(task, inst),
		},
		Spec: cronSpec,
	}
}

// buildTaskPodSpec constructs the PodSpec shared between Job and CronJob templates.
// Redroid runs as a native sidecar (initContainer with restartPolicy=Always, K8s >= 1.29)
// and integration containers run alongside it, sharing localhost networking.
func (r *RedroidTaskReconciler) buildTaskPodSpec(
	task *redroidv1alpha1.RedroidTask,
	inst *redroidv1alpha1.RedroidInstance,
) corev1.PodSpec {
	indexStr := fmt.Sprintf("%d", inst.Spec.Index)
	hostPathType := corev1.HostPathDirectoryOrCreate
	privileged := true
	restartAlways := corev1.ContainerRestartPolicyAlways
	port := instanceADBPort(inst)

	redroidArgs := buildRedroidArgs(inst.Spec)

	// Merge image pull secrets from instance and task.
	pullSecrets := append([]corev1.LocalObjectReference{}, inst.Spec.ImagePullSecrets...)
	pullSecrets = append(pullSecrets, task.Spec.ImagePullSecrets...)

	instPullPolicy := inst.Spec.ImagePullPolicy
	if instPullPolicy == "" {
		instPullPolicy = corev1.PullIfNotPresent
	}

	// Base volumes always present.
	volumes := []corev1.Volume{
		{
			Name: "data-base",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: inst.Spec.SharedDataPVC,
				},
			},
		},
		{
			Name: "data-diff",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: inst.Spec.DiffDataPVC,
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

	// Collect config volumes from all integrations (deduplicated by ConfigMap name).
	seenVolumes := map[string]struct{}{"data-base": {}, "data-diff": {}, "dev-dri": {}}
	for _, integ := range task.Spec.Integrations {
		for _, cfg := range integ.Configs {
			// Use the ConfigMap name as the volume name so that different integrations
			// sharing the same ConfigMap reuse a single volume, and different ConfigMaps
			// within the same integration each get their own volume.
			volName := "cm-" + cfg.ConfigMapName
			if _, exists := seenVolumes[volName]; !exists {
				seenVolumes[volName] = struct{}{}
				volumes = append(volumes, corev1.Volume{
					Name: volName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: cfg.ConfigMapName,
							},
						},
					},
				})
			}
		}
	}

	// Build integration containers.
	integrationContainers := make([]corev1.Container, 0, len(task.Spec.Integrations))
	for _, integ := range task.Spec.Integrations {
		pullPolicy := integ.ImagePullPolicy
		if pullPolicy == "" {
			pullPolicy = corev1.PullAlways
		}

		env := []corev1.EnvVar{
			{Name: "ADB_ADDRESS", Value: fmt.Sprintf("127.0.0.1:%d", port)},
			{Name: "INSTANCE_INDEX", Value: indexStr},
		}
		env = append(env, integ.Env...)

		mounts := append([]corev1.VolumeMount{}, integ.VolumeMounts...)
		for _, cfg := range integ.Configs {
			volName := "cm-" + cfg.ConfigMapName
			mounts = append(mounts, corev1.VolumeMount{
				Name:      volName,
				MountPath: cfg.MountPath,
				SubPath:   cfg.Key,
				ReadOnly:  true,
			})
		}

		// Prepend the ADB readiness check.
		cmd := []string{"/bin/sh", "-c"}
		args := []string{adbReadyScript + "\nexec " + shellJoin(integ.Command, integ.Args)}
		if len(integ.Command) == 0 && len(integ.Args) == 0 {
			args = []string{adbReadyScript}
		}

		c := corev1.Container{
			Name:            integ.Name,
			Image:           integ.Image,
			ImagePullPolicy: pullPolicy,
			Command:         cmd,
			Args:            args,
			WorkingDir:      integ.WorkingDir,
			Env:             env,
			VolumeMounts:    mounts,
			Resources:       integ.Resources,
		}
		if integ.SecurityContext != nil {
			c.SecurityContext = integ.SecurityContext
		}
		integrationContainers = append(integrationContainers, c)
	}

	// Apply ServiceAccountName from the first integration that sets it.
	// All containers in a Pod share one ServiceAccount.
	serviceAccountName := ""
	for _, integ := range task.Spec.Integrations {
		if integ.ServiceAccountName != "" {
			serviceAccountName = integ.ServiceAccountName
			break
		}
	}

	sidecarRestartPolicy := restartAlways
	return corev1.PodSpec{
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: serviceAccountName,
		NodeSelector:       inst.Spec.NodeSelector,
		Tolerations:        inst.Spec.Tolerations,
		Affinity:           inst.Spec.Affinity,
		ImagePullSecrets:   pullSecrets,
		// Redroid runs as a native sidecar (K8s >= 1.29):
		// it starts before containers and is killed after all containers exit.
		InitContainers: []corev1.Container{
			{
				Name:            "redroid",
				Image:           inst.Spec.Image,
				ImagePullPolicy: instPullPolicy,
				Args:            redroidArgs,
				Resources:       inst.Spec.Resources,
				RestartPolicy:   &sidecarRestartPolicy,
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
				Ports: []corev1.ContainerPort{
					{ContainerPort: port, Protocol: corev1.ProtocolTCP},
				},
				Env: append([]corev1.EnvVar{
					{Name: "INSTANCE_INDEX", Value: indexStr},
				}, inst.Spec.ExtraEnv...),
				VolumeMounts: []corev1.VolumeMount{
					{Name: "data-base", MountPath: "/data-base"},
					{Name: "data-diff", MountPath: "/data-diff/" + indexStr, SubPath: indexStr},
					{Name: "dev-dri", MountPath: "/dev/dri"},
				},
			},
		},
		Containers: integrationContainers,
		Volumes:    volumes,
	}
}

// shellJoin serializes a command + args list for embedding in a shell -c string.
func shellJoin(command, args []string) string {
	all := append(command, args...)
	if len(all) == 0 {
		return ""
	}
	result := ""
	for _, a := range all {
		result += fmt.Sprintf("%q ", a)
	}
	return result
}

func taskLabels(task *redroidv1alpha1.RedroidTask, inst *redroidv1alpha1.RedroidInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "redroid-operator",
		"redroid.io/task":              task.Name,
		"redroid.io/instance":          inst.Name,
	}
}

func isJobFinished(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) &&
			c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func ptr[T any](v T) *T { return &v }

// IsJobFinished is the exported variant of isJobFinished for use in tests.
func IsJobFinished(job *batchv1.Job) bool { return isJobFinished(job) }

// SetupWithManager registers the controller with the Manager.
func (r *RedroidTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redroidv1alpha1.RedroidTask{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
