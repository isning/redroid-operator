package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
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
	taskFinalizer = "redroid.isning.moe/task-finalizer"

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
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidtasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidtasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidtasks/finalizers,verbs=update
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidinstances,verbs=get;list;watch
// +kubebuilder:rbac:groups=redroid.isning.moe,resources=redroidinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
type RedroidTaskReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
	pairs, err := r.resolveInstances(ctx, task)
	if err != nil {
		return ctrl.Result{}, err
	}

	if task.Spec.Schedule == "" {
		// One-shot: create one Job per instance.
		return r.reconcileJobs(ctx, task, pairs, logger)
	}
	// Scheduled: create one CronJob per instance.
	return r.reconcileCronJobs(ctx, task, pairs, logger)
}

// resolveInstances fetches all RedroidInstance resources referenced by the task, returning explicit (ref, inst) pairs.
func (r *RedroidTaskReconciler) resolveInstances(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
) ([]InstancePair, error) {
	result := make([]InstancePair, 0, len(task.Spec.Instances))
	seen := make(map[string]bool, len(task.Spec.Instances))
	for i := range task.Spec.Instances {
		ref := &task.Spec.Instances[i]
		if seen[ref.Name] {
			return nil, fmt.Errorf("duplicate instance name %q in spec.instances", ref.Name)
		}
		seen[ref.Name] = true
		inst := &redroidv1alpha1.RedroidInstance{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      ref.Name,
			Namespace: task.Namespace,
		}, inst); err != nil {
			return nil, fmt.Errorf("resolve instance %q: %w", ref.Name, err)
		}
		result = append(result, InstancePair{Ref: ref, Inst: inst})
	}
	return result, nil
}

// InstancePair holds a reference to both the InstanceRef and the resolved RedroidInstance.
type InstancePair struct {
	Ref  *redroidv1alpha1.InstanceRef
	Inst *redroidv1alpha1.RedroidInstance
}

// reconcileJobs ensures one Job per instance exists for a one-shot task.
func (r *RedroidTaskReconciler) reconcileJobs(
	ctx context.Context,
	task *redroidv1alpha1.RedroidTask,
	pairs []InstancePair,
	logger interface{ Info(string, ...interface{}) },
) (ctrl.Result, error) {
	activeJobs := []string{}
	requeueAfter := time.Duration(0)

	for _, pair := range pairs {
		inst := pair.Inst
		ref := pair.Ref
		jobName := fmt.Sprintf("%s-%s", task.Name, inst.Name)
		job := &batchv1.Job{}
		err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: task.Namespace}, job)

		if apierrors.IsNotFound(err) {
			if task.Spec.SuspendInstance {
				if inst.Status.Suspended == nil {
					if setErr := r.setInstanceSuspended(ctx, inst,
						"task/"+task.Name,
						"reserved for one-shot task "+task.Name,
					); setErr != nil {
						return ctrl.Result{}, setErr
					}
					r.Recorder.Eventf(task, corev1.EventTypeNormal, "SuspendedInstance", "Suspended instance %s for task %s", inst.Name, task.Name)
					logger.Info("set suspended on instance; waiting for Pod to stop",
						"instance", inst.Name)
					requeueAfter = 5 * time.Second
					continue
				}
				if inst.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
					logger.Info("waiting for instance to stop before creating Job",
						"instance", inst.Name, "phase", inst.Status.Phase)
					requeueAfter = 5 * time.Second
					continue
				}
			} else if task.Spec.WakeInstance {
				if inst.Status.Woken == nil {
					if setErr := r.setInstanceWoken(ctx, inst,
						"task/"+task.Name,
						"on-demand wake for one-shot task "+task.Name,
					); setErr != nil {
						return ctrl.Result{}, setErr
					}
					r.Recorder.Eventf(task, corev1.EventTypeNormal, "WokenInstance", "Woken instance %s for task %s", inst.Name, task.Name)
					logger.Info("set woken on instance; waiting for Pod to start",
						"instance", inst.Name)
					requeueAfter = 5 * time.Second
					continue
				}
				if inst.Status.Phase != redroidv1alpha1.RedroidInstanceRunning {
					logger.Info("waiting for instance to start before creating Job",
						"instance", inst.Name, "phase", inst.Status.Phase)
					requeueAfter = 5 * time.Second
					continue
				}
			}
			logger.Info("creating Job for instance", "job", jobName, "instance", inst.Name)
			newJob := r.buildJob(task, ref, inst, jobName)
			if err := controllerutil.SetControllerReference(task, newJob, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, newJob); err != nil {
				r.Recorder.Eventf(task, corev1.EventTypeWarning, "FailedCreateJob", "Failed to create Job %s: %v", jobName, err)
				return ctrl.Result{}, fmt.Errorf("create job %s: %w", jobName, err)
			}
			r.Recorder.Eventf(task, corev1.EventTypeNormal, "CreatedJob", "Created Job %s", jobName)
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
				r.Recorder.Eventf(task, corev1.EventTypeNormal, "ClearedSuspendedInstance", "Cleared suspended on instance %s", inst.Name)
				logger.Info("cleared suspended after Job completion",
					"instance", inst.Name, "job", jobName)
			}
			if task.Spec.WakeInstance && inst.Status.Woken != nil {
				if clrErr := r.clearInstanceWoken(ctx, inst); clrErr != nil {
					return ctrl.Result{}, clrErr
				}
				r.Recorder.Eventf(task, corev1.EventTypeNormal, "ClearedWokenInstance", "Cleared woken on instance %s", inst.Name)
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
	pairs []InstancePair,
	logger interface{ Info(string, ...interface{}) },
) (ctrl.Result, error) {
	for _, pair := range pairs {
		inst := pair.Inst
		ref := pair.Ref
		cronName := fmt.Sprintf("%s-%s", task.Name, inst.Name)
		existing := &batchv1.CronJob{}
		err := r.Get(ctx, types.NamespacedName{Name: cronName, Namespace: task.Namespace}, existing)

		desired := r.buildCronJob(task, ref, inst, cronName)
		if err := controllerutil.SetControllerReference(task, desired, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if apierrors.IsNotFound(err) {
			logger.Info("creating CronJob for instance", "cronjob", cronName, "instance", inst.Name)
			if err := r.Create(ctx, desired); err != nil {
				r.Recorder.Eventf(task, corev1.EventTypeWarning, "FailedCreateCronJob", "Failed to create CronJob %s: %v", cronName, err)
				return ctrl.Result{}, fmt.Errorf("create cronjob %s: %w", cronName, err)
			}
			r.Recorder.Eventf(task, corev1.EventTypeNormal, "CreatedCronJob", "Created CronJob %s", cronName)
			continue
		}
		if err != nil {
			return ctrl.Result{}, err
		}

		// Update mutable fields from task spec.
		patch := client.MergeFrom(existing.DeepCopy())
		existing.Spec.Schedule = desired.Spec.Schedule
		existing.Spec.ConcurrencyPolicy = desired.Spec.ConcurrencyPolicy
		existing.Spec.Suspend = desired.Spec.Suspend
		// Explicitly assign (or clear) pointer fields so removing a value from
		// the task spec propagates correctly (e.g. clearing Timezone sets this to nil).
		existing.Spec.TimeZone = desired.Spec.TimeZone
		existing.Spec.StartingDeadlineSeconds = desired.Spec.StartingDeadlineSeconds
		existing.Spec.SuccessfulJobsHistoryLimit = desired.Spec.SuccessfulJobsHistoryLimit
		existing.Spec.FailedJobsHistoryLimit = desired.Spec.FailedJobsHistoryLimit
		// Sync the full pod template so changes to volumes, containers, images,
		// env vars, and integration args are reconciled on every pass.
		// Replace the entire JobTemplate.Spec so fields like BackoffLimit and
		// ActiveDeadlineSeconds from buildCronJob also propagate.
		existing.Spec.JobTemplate.Spec = desired.Spec.JobTemplate.Spec
		existing.Spec.JobTemplate.ObjectMeta = desired.Spec.JobTemplate.ObjectMeta
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

	// Active condition: one or more Jobs are currently running.
	activeStatus := metav1.ConditionFalse
	activeReason := "NoActiveJobs"
	activeMsg := "No Jobs are currently running."
	if len(activeJobs) > 0 {
		activeStatus = metav1.ConditionTrue
		activeReason = "JobsActive"
		activeMsg = fmt.Sprintf("%d Job(s) running: %s.", len(activeJobs), strings.Join(activeJobs, ", "))
	}
	setCondition(&task.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidTaskConditionActive),
		Status:             activeStatus,
		ObservedGeneration: task.Generation,
		Reason:             activeReason,
		Message:            activeMsg,
	})

	// For one-shot tasks, enumerate owned Jobs to derive Complete and Failed conditions.
	// For scheduled tasks (CronJob-based), we set stable placeholder conditions.
	completeStatus := metav1.ConditionFalse
	completeReason := "NotComplete"
	completeMsg := "Not all Jobs have completed successfully."

	failedStatus := metav1.ConditionFalse
	failedReason := "NoFailedJobs"
	failedMsg := "No Jobs have failed."

	if task.Spec.Schedule == "" {
		// List all Jobs owned by this task.
		var jobList batchv1.JobList
		if err := r.List(ctx, &jobList,
			client.InNamespace(task.Namespace),
			client.MatchingLabels{"redroid.isning.moe/task": task.Name},
		); err != nil {
			return ctrl.Result{}, fmt.Errorf("list jobs for task status: %w", err)
		}

		expected := len(task.Spec.Instances)
		succeeded := 0
		var failedNames []string
		var failedDetails []string

		for i := range jobList.Items {
			job := &jobList.Items[i]
			for _, c := range job.Status.Conditions {
				if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
					succeeded++
				}
				if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
					failedNames = append(failedNames, job.Name)
					detail := fmt.Sprintf("job %q failed", job.Name)
					if c.Reason != "" {
						detail += fmt.Sprintf(" (reason: %s)", c.Reason)
					}
					if c.Message != "" {
						detail += fmt.Sprintf(": %s", c.Message)
					}
					failedDetails = append(failedDetails, detail)
				}
			}
		}

		if len(failedNames) > 0 {
			failedStatus = metav1.ConditionTrue
			failedReason = "JobsFailed"
			failedMsg = strings.Join(failedDetails, "; ")
		}

		if len(activeJobs) == 0 && len(failedNames) == 0 && succeeded >= expected && expected > 0 {
			completeStatus = metav1.ConditionTrue
			completeReason = "AllJobsSucceeded"
			if task.Status.LastSuccessfulTime != nil {
				completeMsg = fmt.Sprintf("All %d Job(s) completed successfully at %s.",
					succeeded, task.Status.LastSuccessfulTime.Time.UTC().Format(time.RFC3339))
			} else {
				completeMsg = fmt.Sprintf("All %d Job(s) completed successfully.", succeeded)
			}
		} else if len(activeJobs) == 0 && len(failedNames) == 0 && expected > 0 {
			completeMsg = fmt.Sprintf("%d/%d Job(s) have completed successfully.", succeeded, expected)
		} else if expected > 0 {
			completeMsg = fmt.Sprintf("%d/%d Job(s) have completed successfully; %d active.",
				succeeded, expected, len(activeJobs))
		}
	} else {
		// Scheduled task: conditions are not meaningful per-run; report stable state.
		completeReason = "Scheduled"
		completeMsg = "Task is CronJob-based; per-run completion is tracked via status.lastSuccessfulTime."
		failedReason = "Scheduled"
		failedMsg = "Task is CronJob-based; per-run failures are visible in the owned Job history."
	}

	setCondition(&task.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidTaskConditionComplete),
		Status:             completeStatus,
		ObservedGeneration: task.Generation,
		Reason:             completeReason,
		Message:            completeMsg,
	})
	setCondition(&task.Status.Conditions, metav1.Condition{
		Type:               string(redroidv1alpha1.RedroidTaskConditionFailed),
		Status:             failedStatus,
		ObservedGeneration: task.Generation,
		Reason:             failedReason,
		Message:            failedMsg,
	})

	return ctrl.Result{}, r.Status().Patch(ctx, task, patch)
}

// buildJob constructs a Job manifest that runs all integrations against a single Redroid instance.
func (r *RedroidTaskReconciler) buildJob(
	task *redroidv1alpha1.RedroidTask,
	ref *redroidv1alpha1.InstanceRef,
	inst *redroidv1alpha1.RedroidInstance,
	name string,
) *batchv1.Job {
	podSpec := r.buildTaskPodSpec(task, inst, ref)

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
	ref *redroidv1alpha1.InstanceRef,
	inst *redroidv1alpha1.RedroidInstance,
	name string,
) *batchv1.CronJob {
	podSpec := r.buildTaskPodSpec(task, inst, ref)

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
	ref *redroidv1alpha1.InstanceRef,
) corev1.PodSpec {
	// Nil-safe guard: callers may pass nil when there is no per-instance ref.
	refSafe := ref
	if refSafe == nil {
		refSafe = &redroidv1alpha1.InstanceRef{}
	}

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

	// Build the final volume list: reserved base volumes first, then config-map
	// volumes auto-generated from integrations, then task-level extras, then
	// per-instance overrides (instance may replace task-level but not reserved/generated).
	volumes = mergeVolumes(volumes, task.Spec.Integrations, task.Spec.Volumes, refSafe.Volumes)

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

		// Merge volume mounts: integration-level, config-derived, and per-instance
		// overrides — deduplicated by MountPath, sorted for a deterministic pod spec.
		mounts := mergeVolumeMounts(integ.VolumeMounts, integ.Configs, refSafe.VolumeMounts)

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

	sidecarRestartPolicy := restartAlways
	return corev1.PodSpec{
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: task.Spec.ServiceAccountName,
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
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"/system/bin/sh",
								"-c",
								`test "1" = "$(/system/bin/getprop sys.boot_completed)"`,
							},
						},
					},
					InitialDelaySeconds: 5,
				},
			},
		},
		Containers: integrationContainers,
		Volumes:    volumes,
	}
}

// Volume origin constants used by mergeVolumes to enforce override precedence.
const (
	volOriginReserved  = "reserved"
	volOriginGenerated = "generated"
	volOriginTask      = "task"
	volOriginInstance  = "instance"
)

// mergeVolumes builds the final volume list for a task pod.
// baseVolumes contains the reserved entries (data-base, data-diff, dev-dri) which
// are never overrideable. ConfigMap volumes are auto-generated from integrations
// (cm-* prefix, deduplicated by ConfigMap name). Task-level volumes are appended
// next; instance-level volumes may replace task-level ones but never reserved or
// generated entries.
func mergeVolumes(
	baseVolumes []corev1.Volume,
	integrations []redroidv1alpha1.IntegrationSpec,
	taskVolumes []corev1.Volume,
	instanceVolumes []corev1.Volume,
) []corev1.Volume {
	volumes := append([]corev1.Volume{}, baseVolumes...)
	seen := make(map[string]string, len(baseVolumes))
	volIndex := make(map[string]int, len(baseVolumes)) // name → index in volumes
	for i, v := range baseVolumes {
		seen[v.Name] = volOriginReserved
		volIndex[v.Name] = i
	}

	for _, integ := range integrations {
		for _, cfg := range integ.Configs {
			// Use the ConfigMap name as the volume name so that different integrations
			// sharing the same ConfigMap reuse a single volume.
			volName := ConfigMapVolumeName(cfg.ConfigMapName)
			if _, exists := seen[volName]; !exists {
				volIndex[volName] = len(volumes)
				seen[volName] = volOriginGenerated
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

	for i := range taskVolumes {
		v := &taskVolumes[i]
		if _, exists := seen[v.Name]; !exists {
			volIndex[v.Name] = len(volumes)
			seen[v.Name] = volOriginTask
			volumes = append(volumes, *v)
		}
	}

	for i := range instanceVolumes {
		v := &instanceVolumes[i]
		origin, exists := seen[v.Name]
		if !exists {
			volIndex[v.Name] = len(volumes)
			seen[v.Name] = volOriginInstance
			volumes = append(volumes, *v)
		} else if origin == volOriginTask {
			// O(1) replacement using the index map; reserved/generated entries are skipped.
			volumes[volIndex[v.Name]] = *v
		}
	}
	return volumes
}

// mergeVolumeMounts builds the deduplicated, deterministically sorted VolumeMount
// list for one integration container. integMounts and config-derived mounts form
// the base; instanceMounts can override any entry keyed by MountPath.
func mergeVolumeMounts(
	integMounts []corev1.VolumeMount,
	configs []redroidv1alpha1.ConfigFile,
	instanceMounts []corev1.VolumeMount,
) []corev1.VolumeMount {
	byPath := make(map[string]corev1.VolumeMount, len(integMounts)+len(configs)+len(instanceMounts))
	for _, m := range integMounts {
		byPath[m.MountPath] = m
	}
	for _, cfg := range configs {
		byPath[cfg.MountPath] = corev1.VolumeMount{
			Name:      ConfigMapVolumeName(cfg.ConfigMapName),
			MountPath: cfg.MountPath,
			SubPath:   cfg.Key,
			ReadOnly:  true,
		}
	}
	for _, m := range instanceMounts {
		byPath[m.MountPath] = m
	}
	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	mounts := make([]corev1.VolumeMount, 0, len(paths))
	for _, p := range paths {
		mounts = append(mounts, byPath[p])
	}
	return mounts
}

// ConfigMapVolumeName returns a DNS-1123-label-safe Kubernetes volume name for
// the given ConfigMap name. It lowercases the input, replaces dots and
// underscores with hyphens, strips any remaining non-alphanumeric/non-hyphen
// characters, and prepends the "cm-" prefix. A deterministic 8-character
// SHA-256 suffix (derived from the original name) is always appended so that
// normalized forms that collide (e.g. "foo.bar" vs "foo-bar") remain unique.
// The base is truncated if necessary so that prefix + base + suffix <= 63.
func ConfigMapVolumeName(configMapName string) string {
	const prefix = "cm-"
	const maxLen = 63

	var b strings.Builder
	for _, r := range strings.ToLower(configMapName) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == '.', r == '_':
			b.WriteRune('-')
			// other characters are stripped
		}
	}
	base := b.String()

	// Always append a deterministic hash so that distinct names whose
	// normalized forms collide still produce distinct volume names.
	hash := sha256.Sum256([]byte(configMapName))
	suffix := "-" + hex.EncodeToString(hash[:])[:8] // 9 chars including leading dash

	allowed := maxLen - len(prefix) - len(suffix)
	if len(base) > allowed {
		base = base[:allowed]
	}
	return prefix + base + suffix
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
		"redroid.isning.moe/task":      task.Name,
		"redroid.isning.moe/instance":  inst.Name,
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
