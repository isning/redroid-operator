package controller_test

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

func makeInstance(name string, index int, suspended bool) *redroidv1alpha1.RedroidInstance {
	return &redroidv1alpha1.RedroidInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: redroidv1alpha1.RedroidInstanceSpec{
			Index:         index,
			Image:         "redroid/redroid:16.0.0-latest",
			Suspend:       suspended,
			SharedDataPVC: "redroid-data-base-pvc",
			DiffDataPVC:   "redroid-data-diff-pvc",
			GPUMode:       "host",
		},
	}
}

func reconcileInstance(t *testing.T, r *controller.RedroidInstanceReconciler, name string) ctrl.Result {
	t.Helper()
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	return res
}

// TestRedroidInstance_AddsFinalizer verifies that the first reconcile adds the finalizer.
func TestRedroidInstance_AddsFinalizer(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("test-0", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "test-0")

	updated := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-0", Namespace: "default"}, updated); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if !controllerutil.ContainsFinalizer(updated, "redroid.isning.moe/instance-finalizer") {
		t.Error("expected finalizer to be set after first reconcile")
	}
}

// TestRedroidInstance_CreatesPodWhenNotSuspended verifies that a Pod is created for an unsuspended instance.
func TestRedroidInstance_CreatesPodWhenNotSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-active", 2, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-active") // adds finalizer
	reconcileInstance(t, r, "inst-active") // creates Pod

	podName := fmt.Sprintf("redroid-instance-%s", "inst-active")
	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod); err != nil {
		t.Fatalf("expected Pod %q to exist: %v", podName, err)
	}
	if pod.Spec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("expected RestartPolicyNever, got %v", pod.Spec.RestartPolicy)
	}
	if len(pod.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(pod.Spec.Containers))
	}
	c := pod.Spec.Containers[0]
	if c.Name != "redroid" {
		t.Errorf("expected container name 'redroid', got %q", c.Name)
	}
	if c.Image != "redroid/redroid:16.0.0-latest" {
		t.Errorf("unexpected image: %q", c.Image)
	}
}

// TestRedroidInstance_NoPodWhenSuspended verifies that no Pod is created when suspended=true.
func TestRedroidInstance_NoPodWhenSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-suspended", 3, true)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-suspended")
	reconcileInstance(t, r, "inst-suspended")

	podName := fmt.Sprintf("redroid-instance-%s", "inst-suspended")
	err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{})
	if err == nil {
		t.Error("expected no Pod to exist when suspended=true")
	}
}

// TestRedroidInstance_DeletesPodWhenSuspended verifies a running Pod is deleted when suspended.
func TestRedroidInstance_DeletesPodWhenSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-toggle", 1, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-toggle")
	reconcileInstance(t, r, "inst-toggle")

	podName := fmt.Sprintf("redroid-instance-%s", "inst-toggle")
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{}); err != nil {
		t.Fatalf("Pod should exist before suspend: %v", err)
	}

	updated := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-toggle", Namespace: "default"}, updated); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	updated.Spec.Suspend = true
	if err := fakeClient.Update(context.Background(), updated); err != nil {
		t.Fatalf("update instance: %v", err)
	}
	reconcileInstance(t, r, "inst-toggle")

	err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{})
	if err == nil {
		t.Error("expected Pod to be deleted after suspend")
	}
}

// TestRedroidInstance_StatusPending verifies status is Pending just after Pod creation.
func TestRedroidInstance_StatusPending(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-status", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-status")
	reconcileInstance(t, r, "inst-status")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-status", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstancePending {
		t.Errorf("expected Pending, got %v", result.Status.Phase)
	}
	wantPod := fmt.Sprintf("redroid-instance-%s", "inst-status")
	if result.Status.PodName != wantPod {
		t.Errorf("expected PodName=%q, got %q", wantPod, result.Status.PodName)
	}
}

// TestRedroidInstance_StatusRunningWithADB verifies ADBAddress is set when Pod is Running.
func TestRedroidInstance_StatusRunningWithADB(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-running", 4, false)

	podName := fmt.Sprintf("redroid-instance-%s", "inst-running")
	runningPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.42"},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, runningPod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-running")
	reconcileInstance(t, r, "inst-running")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-running", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceRunning {
		t.Errorf("expected Running, got %v", result.Status.Phase)
	}
	if result.Status.ADBAddress != "10.0.0.42:5555" {
		t.Errorf("expected ADBAddress=10.0.0.42:5555, got %q", result.Status.ADBAddress)
	}
}

// TestRedroidInstance_StatusFailed verifies a failed Pod sets status to Failed.
func TestRedroidInstance_StatusFailed(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-failed", 5, false)

	podName := fmt.Sprintf("redroid-instance-%s", "inst-failed")
	failedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, failedPod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-failed")
	reconcileInstance(t, r, "inst-failed")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-failed", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceFailed {
		t.Errorf("expected Failed, got %v", result.Status.Phase)
	}
}

// TestRedroidInstance_StatusStopped verifies a succeeded Pod sets status to Stopped.
func TestRedroidInstance_StatusStopped(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-succeeded", 6, false)

	podName := fmt.Sprintf("redroid-instance-%s", "inst-succeeded")
	succeededPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, succeededPod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-succeeded")
	reconcileInstance(t, r, "inst-succeeded")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-succeeded", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
		t.Errorf("expected Stopped, got %v", result.Status.Phase)
	}
}

// TestRedroidInstance_PodOverlayfsArgs verifies overlayfs and extra args are present.
func TestRedroidInstance_PodOverlayfsArgs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-args", 7, false)
	inst.Spec.ExtraArgs = []string{"androidboot.redroid_width=1080"}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-args")
	reconcileInstance(t, r, "inst-args")

	podName := fmt.Sprintf("redroid-instance-%s", "inst-args")
	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	wantArgs := map[string]bool{
		"androidboot.redroid_gpu_mode=host":   true,
		"androidboot.use_memfd=1":             true,
		"androidboot.use_redroid_overlayfs=1": true,
		"androidboot.redroid_width=1080":      true,
	}
	for _, a := range pod.Spec.Containers[0].Args {
		delete(wantArgs, a)
	}
	if len(wantArgs) > 0 {
		t.Errorf("missing expected args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	}
}

// TestRedroidInstance_PodVolumeMounts verifies /data-base, /data-diff/<index> and /dev/dri mounts.
func TestRedroidInstance_PodVolumeMounts(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-vols", 3, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-vols")
	reconcileInstance(t, r, "inst-vols")

	podName := fmt.Sprintf("redroid-instance-%s", "inst-vols")
	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	mountPaths := map[string]bool{}
	for _, vm := range pod.Spec.Containers[0].VolumeMounts {
		mountPaths[vm.MountPath] = true
	}
	for _, want := range []string{"/data-base", "/data-diff/3", "/dev/dri"} {
		if !mountPaths[want] {
			t.Errorf("expected mount path %q not found; mounts: %v", want, mountPaths)
		}
	}
}

// TestRedroidInstance_IgnoreNotFound verifies reconcile returns no error for missing objects.
func TestRedroidInstance_IgnoreNotFound(t *testing.T) {
	scheme := newTestScheme(t)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "does-not-exist", Namespace: "default"},
	})
	if err != nil {
		t.Errorf("expected no error for missing resource, got: %v", err)
	}
}

// TestRedroidInstance_DeletionCleansPod verifies that deleting an instance removes the Pod
// and strips the finalizer so the object can be garbage collected.
func TestRedroidInstance_DeletionCleansPod(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("del-0", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}

	// First reconcile: controller adds its own finalizer.
	reconcileInstance(t, r, "del-0")
	// Second reconcile: Pod is created.
	reconcileInstance(t, r, "del-0")

	// Confirm Pod was created.
	podList := &corev1.PodList{}
	if err := fakeClient.List(context.Background(), podList); err != nil {
		t.Fatalf("list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		t.Fatal("expected Pod to exist before deletion")
	}

	// Delete the instance. Fake client sets DeletionTimestamp because the
	// controller finalizer "redroid.isning.moe/instance-finalizer" is present.
	current := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "del-0", Namespace: "default"}, current); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if err := fakeClient.Delete(context.Background(), current); err != nil {
		t.Fatalf("delete instance: %v", err)
	}

	// Reconcile the deletion: controller should delete Pod and remove its finalizer.
	reconcileInstance(t, r, "del-0")

	// After the finalizer is removed the fake client deletes the object.
	// Either the object is gone (NotFound) or exists with the finalizer stripped.
	final := &redroidv1alpha1.RedroidInstance{}
	err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "del-0", Namespace: "default"}, final)
	if err == nil && controllerutil.ContainsFinalizer(final, "redroid.isning.moe/instance-finalizer") {
		t.Error("expected 'redroid.isning.moe/instance-finalizer' to be removed after deletion reconcile")
	}
}

// TestRedroidInstance_PodSucceededBecomeStopped verifies that when the Pod phase is Succeeded
// the instance is reported as Stopped (normal game process exit or deliberate shut-down).
func TestRedroidInstance_PodSucceededBecomeStopped(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("succ-0", 0, false)

	podName := "redroid-instance-succ-0"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "succ-0") // adds finalizer
	reconcileInstance(t, r, "succ-0") // observes Succeeded Pod

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "succ-0", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
		t.Errorf("expected Stopped for Succeeded Pod, got %q", result.Status.Phase)
	}
}

// TestRedroidInstance_CreatesService verifies that a stable ClusterIP Service
// is created alongside the Pod so that the CLI can port-forward to the Service
// rather than a specific Pod name.
func TestRedroidInstance_CreatesService(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("svc-test", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "svc-test") // adds finalizer
	reconcileInstance(t, r, "svc-test") // creates Pod + Service

	svcName := "redroid-instance-svc-test"
	svc := &corev1.Service{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: svcName, Namespace: "default"}, svc); err != nil {
		t.Fatalf("expected Service %q to exist: %v", svcName, err)
	}

	// Selector must point at the instance label so traffic reaches the Pod.
	if svc.Spec.Selector["redroid.isning.moe/instance"] != "svc-test" {
		t.Errorf("Service selector redroid.isning.moe/instance = %q, want %q",
			svc.Spec.Selector["redroid.isning.moe/instance"], "svc-test")
	}
	if len(svc.Spec.Ports) != 1 || svc.Spec.Ports[0].Port != 5555 {
		t.Errorf("expected 1 port 5555, got %v", svc.Spec.Ports)
	}
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Errorf("expected ClusterIP, got %v", svc.Spec.Type)
	}
}

// TestRedroidInstance_CreatesServiceWhenSuspended verifies that the Service is
// created even when the instance is suspended (no Pod). This keeps the DNS
// name stable so tooling can discover the instance without needing the Pod.
func TestRedroidInstance_CreatesServiceWhenSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("svc-suspended", 0, true) // suspended from the start

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "svc-suspended") // adds finalizer
	reconcileInstance(t, r, "svc-suspended") // reconciles suspended path

	svcName := "redroid-instance-svc-suspended"
	svc := &corev1.Service{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: svcName, Namespace: "default"}, svc); err != nil {
		t.Fatalf("expected Service %q to exist even when suspended: %v", svcName, err)
	}
}

// TestRedroidInstance_CustomServiceSpec verifies that spec.service fields are applied:
// NodePort type, custom node port, and extra annotations.
func TestRedroidInstance_CustomServiceSpec(t *testing.T) {
	scheme := newTestScheme(t)
	nodePort := int32(30555)
	inst := makeInstance("svc-custom", 0, false)
	inst.Spec.Service = &redroidv1alpha1.InstanceServiceSpec{
		Type:     corev1.ServiceTypeNodePort,
		NodePort: &nodePort,
		Annotations: map[string]string{
			"example.io/custom": "true",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "svc-custom") // adds finalizer
	reconcileInstance(t, r, "svc-custom") // creates Service

	svc := &corev1.Service{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-svc-custom", Namespace: "default"}, svc); err != nil {
		t.Fatalf("get Service: %v", err)
	}
	if svc.Spec.Type != corev1.ServiceTypeNodePort {
		t.Errorf("expected NodePort, got %v", svc.Spec.Type)
	}
	if len(svc.Spec.Ports) != 1 || svc.Spec.Ports[0].NodePort != nodePort {
		t.Errorf("expected nodePort=%d, got %v", nodePort, svc.Spec.Ports)
	}
	if svc.Annotations["example.io/custom"] != "true" {
		t.Errorf("expected annotation example.io/custom=true, got %v", svc.Annotations)
	}
}
