package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceRef selects a RedroidInstance by name to include in a task run.
type InstanceRef struct {
	// Name is the RedroidInstance name in the same namespace.
	Name string `json:"name"`

	// Volumes adds additional volumes specific to this instance's Job Pod.
	// These are merged with task-level spec.volumes; an instance volume overrides
	// only user-defined task-level volumes with the same name. Reserved volumes
	// (data-base, data-diff, dev-dri) and controller-generated ConfigMap volumes
	// (cm-* prefix) are never overrideable — those retain precedence regardless.
	// +optional
	// +listType=map
	// +listMapKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// VolumeMounts adds extra volume mounts into every integration container
	// for this instance only. Use together with instance-level Volumes to
	// mount instance-specific ConfigMaps, Secrets, etc.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// ConfigFile mounts a single key from a ConfigMap as a file inside the integration container.
type ConfigFile struct {
	// ConfigMapName is the name of the ConfigMap in the same namespace.
	ConfigMapName string `json:"configMapName"`
	// Key is the ConfigMap key to mount.
	Key string `json:"key"`
	// MountPath is the absolute path inside the container where the file is placed.
	MountPath string `json:"mountPath"`
}

// IntegrationSpec describes a tool container that runs against the Redroid ADB.
// ADB address is injected as ADB_ADDRESS env var; instance index as INSTANCE_INDEX.
type IntegrationSpec struct {
	// Name is a unique identifier for this integration within the task.
	Name string `json:"name"`

	// Image is the container image for this tool.
	Image string `json:"image"`

	// ImagePullPolicy overrides the image pull policy. Defaults to Always.
	// +optional
	// +kubebuilder:default="Always"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Command overrides the container entrypoint.
	// +optional
	Command []string `json:"command,omitempty"`

	// Args is passed to the container command.
	// +optional
	Args []string `json:"args,omitempty"`

	// WorkingDir sets the current working directory inside the container.
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// Env provides additional environment variables, merged after ADB_ADDRESS/INSTANCE_INDEX.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Configs mounts ConfigMap keys as files inside this container.
	// +optional
	Configs []ConfigFile `json:"configs,omitempty"`

	// VolumeMounts adds extra volume mounts into the integration container.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// SecurityContext sets per-container security options.
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// Resources optionally sets CPU/memory limits on this container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// RedroidTaskSpec defines the desired state of RedroidTask.
// +kubebuilder:validation:XValidation:rule="!(has(self.suspendInstance) && self.suspendInstance == true && has(self.wakeInstance) && self.wakeInstance == true)",message="suspendInstance and wakeInstance are mutually exclusive"
type RedroidTaskSpec struct {
	// Instances lists the RedroidInstance resources that this task targets.
	// Each instance runs its own Job/CronJob execution, inheriting
	// the instance's image, overlayfs PVCs, and GPU settings.
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=name
	Instances []InstanceRef `json:"instances"`

	// Schedule is a Cron expression for recurring execution (e.g. "0 4 * * *").
	// If empty, the task is one-shot and runs immediately upon creation.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Suspend pauses CronJob execution without deleting it. Ignored for one-shot tasks.
	// Follows the same semantics as CronJob.spec.suspend.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// Timezone is the IANA timezone name for the CronJob schedule (e.g. "Asia/Shanghai").
	// Requires Kubernetes >= 1.27. Ignored for one-shot tasks.
	// +optional
	Timezone string `json:"timezone,omitempty"`

	// StartingDeadlineSeconds is the optional deadline (seconds) for starting a CronJob
	// if it misses scheduled time. Ignored for one-shot tasks.
	// +optional
	// +kubebuilder:validation:Minimum=0
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// BackoffLimit specifies the number of retries before marking the Job as failed.
	// Defaults to 0 (no retries).
	// +optional
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// SuspendInstance temporarily stops the referenced RedroidInstance Pod while the
	// one-shot Job runs, then automatically restores it on Job completion/failure.
	// This prevents overlayfs conflicts when both the instance Pod and the task Job
	// would otherwise write to the same /data-diff volume simultaneously.
	//
	// Mechanism: the task controller sets status.suspended on each instance
	// before creating the Job, waits until the instance Pod is stopped (phase=Stopped),
	// then creates the Job. After the Job finishes the controller clears the field.
	//
	// This field is ignored for CronJob-based (scheduled) tasks.
	// Mutually exclusive with WakeInstance.
	// +optional
	SuspendInstance bool `json:"suspendInstance,omitempty"`

	// WakeInstance temporarily starts the referenced RedroidInstance Pod while the
	// one-shot Job runs, then restores the original suspended state on Job completion/failure.
	// Use this for on-demand instances that are normally stopped (spec.suspend: true).
	//
	// Mechanism: the task controller sets status.woken on each instance before creating the
	// Job, waits until the instance Pod reaches phase=Running, then creates the Job.
	// After the Job finishes the controller clears status.woken, allowing spec.suspend to
	// take effect again.
	//
	// This field is ignored for CronJob-based (scheduled) tasks.
	// Mutually exclusive with SuspendInstance.
	// +optional
	WakeInstance bool `json:"wakeInstance,omitempty"`

	// ActiveDeadlineSeconds limits the duration of each Job in seconds.
	// Jobs exceeding this limit are terminated. 0 means no limit.
	// +optional
	// +kubebuilder:validation:Minimum=1
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// TTLSecondsAfterFinished automatically removes completed one-shot Jobs after
	// this many seconds. Ignored for scheduled tasks.
	// +optional
	// +kubebuilder:validation:Minimum=0
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// Parallelism controls how many instance Pods run concurrently.
	// Defaults to the number of Instances (run all in parallel).
	// +optional
	// +kubebuilder:validation:Minimum=1
	Parallelism *int32 `json:"parallelism,omitempty"`

	// Integrations is the ordered list of tool containers executed per instance.
	// They run as regular containers alongside the redroid sidecar, sharing localhost.
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=name
	Integrations []IntegrationSpec `json:"integrations"`

	// Volumes defines additional volumes to attach to the Job Pod.
	// Use this together with integration VolumeMounts to mount arbitrary volume
	// sources (Secrets, projected ConfigMaps, emptyDir, etc.) that are not covered
	// by the per-key Configs shorthand.
	// +optional
	// +listType=map
	// +listMapKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use for all Pods
	// created by this task. Applies at the PodSpec level and is shared by every
	// integration container. If empty, the namespace default ServiceAccount is used.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// ImagePullSecrets applies to all integration containers in this task.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// SuccessfulJobsHistoryLimit controls how many successful CronJob-spawned Jobs to retain. Defaults to 3.
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=0
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// FailedJobsHistoryLimit controls how many failed CronJob-spawned Jobs to retain. Defaults to 3.
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=0
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

// RedroidTaskConditionType defines well-known condition types for RedroidTask.
type RedroidTaskConditionType string

const (
	// RedroidTaskConditionActive is true when at least one Job is currently running.
	RedroidTaskConditionActive RedroidTaskConditionType = "Active"

	// RedroidTaskConditionComplete is true when all one-shot Jobs have completed successfully.
	RedroidTaskConditionComplete RedroidTaskConditionType = "Complete"

	// RedroidTaskConditionFailed is true when one or more Jobs have failed beyond the backoff limit.
	RedroidTaskConditionFailed RedroidTaskConditionType = "Failed"
)

// RedroidTaskStatus defines the observed state of RedroidTask.
type RedroidTaskStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastScheduleTime is the last time the CronJob was scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// LastSuccessfulTime is the last time a Job completed successfully.
	// +optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// ActiveJobs lists the names of currently running Jobs.
	// +optional
	ActiveJobs []string `json:"activeJobs,omitempty"`

	// Conditions holds detailed status conditions (Active, Complete, Failed).
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=".spec.schedule"
// +kubebuilder:printcolumn:name="Suspend",type=boolean,JSONPath=".spec.suspend"
// +kubebuilder:printcolumn:name="Active",type=integer,JSONPath=".status.activeJobs"
// +kubebuilder:printcolumn:name="LastSchedule",type=date,JSONPath=".status.lastScheduleTime"
// +kubebuilder:printcolumn:name="LastSuccess",type=date,JSONPath=".status.lastSuccessfulTime"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// RedroidTask is the Schema for the redroidtasks API.
// It describes a workload (one-shot or recurring via CronJob) that runs integration
// tool containers against a set of RedroidInstance overlay partitions.
type RedroidTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedroidTaskSpec   `json:"spec,omitempty"`
	Status RedroidTaskStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedroidTaskList contains a list of RedroidTask.
type RedroidTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedroidTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedroidTask{}, &RedroidTaskList{})
}
