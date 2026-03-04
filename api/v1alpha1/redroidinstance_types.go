package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScreenSpec configures the virtual display exposed by the Redroid instance.
type ScreenSpec struct {
	// Width is the screen width in pixels. Defaults to 720.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Width *int32 `json:"width,omitempty"`

	// Height is the screen height in pixels. Defaults to 1280.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Height *int32 `json:"height,omitempty"`

	// DPI is the screen density. Defaults to 320.
	// +optional
	// +kubebuilder:validation:Minimum=1
	DPI *int32 `json:"dpi,omitempty"`

	// FPS is the display frame rate. Defaults to 30 (Android 12+) or 15 (older).
	// +optional
	// +kubebuilder:validation:Minimum=1
	FPS *int32 `json:"fps,omitempty"`
}

// ProxySpec configures HTTP/HTTPS proxy for the Redroid network stack.
type ProxySpec struct {
	// Type selects the proxy mode: static, pac, none, or unassigned.
	// +optional
	// +kubebuilder:validation:Enum=static;pac;none;unassigned
	Type string `json:"type,omitempty"`

	// Host is the proxy server hostname or IP (used with Type=static).
	// +optional
	Host string `json:"host,omitempty"`

	// Port is the proxy server port. Defaults to 3128.
	// +optional
	// +kubebuilder:default=3128
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// ExcludeList is a comma-separated list of hosts bypassing the proxy.
	// +optional
	ExcludeList string `json:"excludeList,omitempty"`

	// PAC is the URL of a proxy auto-config file (used with Type=pac).
	// +optional
	PAC string `json:"pac,omitempty"`
}

// InstanceServiceSpec customises the Kubernetes Service that the operator creates
// to front the ADB port of a RedroidInstance.
type InstanceServiceSpec struct {
	// Type is the Service type. Defaults to ClusterIP.
	// Use NodePort or LoadBalancer to expose the ADB port outside the cluster.
	// +optional
	// +kubebuilder:default="ClusterIP"
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type corev1.ServiceType `json:"type,omitempty"`

	// Annotations are extra annotations merged onto the Service.
	// Useful for cloud-provider-specific behaviour, e.g. AWS NLB, GCP NEG.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// NodePort pins the node port when Type=NodePort.
	// If 0 or unset, Kubernetes auto-assigns a port from the node-port range.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	NodePort *int32 `json:"nodePort,omitempty"`
}

// NetworkSpec configures DNS and proxy settings for the Redroid instance.
type NetworkSpec struct {
	// DNS lists DNS server addresses (up to N servers).
	// They map to androidboot.redroid_net_ndns and androidboot.redroid_net_dns<1..N>.
	// +optional
	DNS []string `json:"dns,omitempty"`

	// Proxy configures HTTP/HTTPS proxy for outbound traffic.
	// +optional
	Proxy *ProxySpec `json:"proxy,omitempty"`
}

// RedroidInstanceSpec defines the desired state of a RedroidInstance.
type RedroidInstanceSpec struct {
	// Index is the overlayfs partition index shared with RedroidTasks.
	// /data-base is shared across all instances; /data-diff/<Index> is this instance's private layer.
	// +kubebuilder:validation:Minimum=0
	Index int `json:"index"`

	// Image is the redroid container image. Defaults to redroid/redroid:16.0.0-latest.
	// +optional
	// +kubebuilder:default="redroid/redroid:16.0.0-latest"
	Image string `json:"image,omitempty"`

	// ImagePullPolicy for the redroid container image. Defaults to IfNotPresent.
	// +optional
	// +kubebuilder:default="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets references Secret resources for pulling private images.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Suspend controls whether the Android instance Pod should be running.
	// Set to false (default) to start the instance; true to stop it without deleting the resource.
	// This follows the same semantics as CronJob.spec.suspend.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// SharedDataPVC is the PVC name for the shared /data-base volume (read-only lower layer).
	// +kubebuilder:default="redroid-data-base-pvc"
	SharedDataPVC string `json:"sharedDataPVC"`

	// DiffDataPVC is the PVC name for the per-instance /data-diff volume (writable upper layer).
	// +kubebuilder:default="redroid-data-diff-pvc"
	DiffDataPVC string `json:"diffDataPVC"`

	// BaseMode makes this instance write directly to SharedDataPVC as /data,
	// bypassing the overlayfs per-instance layer. Use this to initialise the shared
	// base image (install common APKs, first-boot setup, etc.).
	//
	// In base mode:
	//   - SharedDataPVC is mounted at /data (read-write).
	//   - DiffDataPVC is NOT mounted.
	//   - androidboot.use_redroid_overlayfs is set to 0.
	//
	// Typical workflow:
	//   1. Create a RedroidInstance with baseMode: true, suspend: false.
	//   2. ADB-connect and perform initial setup (app installs, configs).
	//   3. Set suspend: true (or delete) once satisfied.
	//   4. Normal instances (baseMode: false) sharing the same SharedDataPVC
	//      will see the initialised state as their read-only lower layer.
	//
	// WARNING: running a base-mode instance concurrently with normal instances
	// that share the same SharedDataPVC may corrupt their overlayfs lower layer.
	// Use spec.suspend or status.suspended on all normal instances first.
	// +optional
	BaseMode bool `json:"baseMode,omitempty"`

	// GPUMode sets androidboot.redroid_gpu_mode. Defaults to "host".
	// +optional
	// +kubebuilder:default="host"
	// +kubebuilder:validation:Enum=host;guest;auto;none
	GPUMode string `json:"gpuMode,omitempty"`

	// GPUNode sets androidboot.redroid_gpu_node (DRM device path).
	// When empty the Redroid runtime auto-detects the GPU node.
	// +optional
	GPUNode string `json:"gpuNode,omitempty"`

	// ADBPort is the ADB TCP port exposed by the container. Defaults to 5555.
	// +optional
	// +kubebuilder:default=5555
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ADBPort *int32 `json:"adbPort,omitempty"`

	// Screen configures the virtual display (width, height, dpi, fps).
	// +optional
	Screen *ScreenSpec `json:"screen,omitempty"`

	// Network configures DNS and proxy settings for the Android instance.
	// +optional
	Network *NetworkSpec `json:"network,omitempty"`

	// ExtraArgs are additional androidboot.* arguments passed to the redroid container.
	// Supports $(VAR_NAME) substitution from ExtraEnv.
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// ExtraEnv defines additional environment variables for the redroid container.
	// Supports valueFrom.secretKeyRef and valueFrom.configMapKeyRef for sensitive params.
	// Can be referenced in ExtraArgs via $(VAR_NAME) syntax.
	// +optional
	ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`

	// NodeSelector constrains which node the Pod is scheduled on.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations allows the Pod to be scheduled on nodes with matching taints.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity provides advanced scheduling constraints.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Resources optionally sets CPU/memory limits on the redroid container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Service customises the Kubernetes Service the operator creates to expose the ADB port.
	// The Service is always created; this field controls its type, annotations, and node port.
	// +optional
	Service *InstanceServiceSpec `json:"service,omitempty"`
}

// RedroidInstanceConditionType defines well-known condition types for RedroidInstance.
type RedroidInstanceConditionType string

const (
	// RedroidInstanceConditionReady is true when the Pod is Running and the ADB address is known.
	RedroidInstanceConditionReady RedroidInstanceConditionType = "Ready"

	// RedroidInstanceConditionScheduled is true once a Pod has been successfully created.
	RedroidInstanceConditionScheduled RedroidInstanceConditionType = "Scheduled"
)

// RedroidInstancePhase describes the lifecycle phase of a RedroidInstance.
// +kubebuilder:validation:Enum=Pending;Running;Stopped;Failed
type RedroidInstancePhase string

// RedroidInstance phase constants.
const (
	RedroidInstancePending RedroidInstancePhase = "Pending"
	RedroidInstanceRunning RedroidInstancePhase = "Running"
	RedroidInstanceStopped RedroidInstancePhase = "Stopped"
	RedroidInstanceFailed  RedroidInstancePhase = "Failed"
)

// SuspendedStatus holds a suspend-override override that stops the instance Pod
// without modifying spec.suspend. Because this field lives in status it is not reconciled by
// GitOps tools such as Flux, preventing config drift.
//
// Set by: operators (manual kubectl patch), or by the RedroidTask controller when
// spec.suspendInstance is true. Clear by patching status.suspended to null,
// or rely on the automatic expiry if Until is set.
type SuspendedStatus struct {
	// Reason is a human-readable explanation for the temporary suspend.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Until is an optional expiry timestamp. The controller automatically clears the
	// temporary suspend (and restarts the Pod) once this time has passed.
	// +optional
	Until *metav1.Time `json:"until,omitempty"`

	// Actor identifies who set the temporary suspend, e.g. "manual", "task/maa-task".
	// +optional
	Actor string `json:"actor,omitempty"`
}

// RedroidInstanceStatus defines the observed state of RedroidInstance.
type RedroidInstanceStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase is the current lifecycle phase.
	// +optional
	Phase RedroidInstancePhase `json:"phase,omitempty"`

	// PodName is the name of the managed Pod, if it exists.
	// +optional
	PodName string `json:"podName,omitempty"`

	// ADBAddress is the in-cluster address (host:port) to reach this instance's ADB.
	// +optional
	ADBAddress string `json:"adbAddress,omitempty"`

	// Conditions holds detailed status conditions (Ready, Scheduled).
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Suspended holds a suspend-override override set by controllers or manual operators.
	// The instance Pod is stopped while this field is non-nil, regardless of spec.suspend.
	// Unlike spec.suspend, this field is not reconciled by GitOps tools (e.g. Flux) so clearing
	// it does not cause config drift.
	//
	// To manually suspend: kubectl patch redroidinstance <name> --subresource=status
	//   --type=merge -p '{"status":{"suspended":{"reason":"maintenance","actor":"manual"}}}'
	// To clear: kubectl patch redroidinstance <name> --subresource=status
	//   --type=merge -p '{"status":{"suspended":null}}'
	// +optional
	Suspended *SuspendedStatus `json:"suspended,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Index",type=integer,JSONPath=".spec.index"
// +kubebuilder:printcolumn:name="Suspend",type=boolean,JSONPath=".spec.suspend"
// +kubebuilder:printcolumn:name="Suspended",type=string,JSONPath=".status.suspended.actor",priority=1
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Pod",type=string,JSONPath=".status.podName"
// +kubebuilder:printcolumn:name="ADB",type=string,JSONPath=".status.adbAddress"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// RedroidInstance is the Schema for the redroidinstances API.
// It represents a single persistent Android container instance backed by overlayfs storage.
type RedroidInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedroidInstanceSpec   `json:"spec,omitempty"`
	Status RedroidInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedroidInstanceList contains a list of RedroidInstance.
type RedroidInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedroidInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedroidInstance{}, &RedroidInstanceList{})
}
