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

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// TestRedroidInstance_ScreenArgs verifies screen resolution args are passed to the redroid container.
func TestRedroidInstance_ScreenArgs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-screen", 0, false)
	width, height, dpi := int32(1080), int32(1920), int32(480)
	inst.Spec.Screen = &redroidv1alpha1.ScreenSpec{Width: &width, Height: &height, DPI: &dpi}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-screen")
	reconcileInstance(t, r, "inst-screen")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-screen", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	wantArgs := map[string]bool{
		"androidboot.redroid_width=1080":  true,
		"androidboot.redroid_height=1920": true,
		"androidboot.redroid_dpi=480":     true,
	}
	for _, a := range pod.Spec.Containers[0].Args {
		delete(wantArgs, a)
	}
	if len(wantArgs) > 0 {
		t.Errorf("missing screen args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	}
}

// TestRedroidInstance_CustomADBPort verifies that a custom ADB port is applied to the Pod and status.
func TestRedroidInstance_CustomADBPort(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-port", 0, false)
	customPort := int32(6666)
	inst.Spec.ADBPort = &customPort

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-port") // adds finalizer
	reconcileInstance(t, r, "inst-port") // creates Pod

	// Check the controller-created Pod uses the custom port.
	podName := "redroid-instance-inst-port"
	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	if len(pod.Spec.Containers) == 0 {
		t.Fatal("expected at least one container")
	}
	found := false
	for _, p := range pod.Spec.Containers[0].Ports {
		if p.ContainerPort == customPort {
			found = true
		}
	}
	if !found {
		t.Errorf("expected container port %d, got: %v", customPort, pod.Spec.Containers[0].Ports)
	}

	// Simulate pod going Running, check ADBAddress uses the custom port.
	pod.Status.Phase = corev1.PodRunning
	pod.Status.PodIP = "10.1.0.5"
	if err := fakeClient.Status().Update(context.Background(), pod); err != nil {
		t.Fatalf("update pod status: %v", err)
	}
	reconcileInstance(t, r, "inst-port")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-port", Namespace: "default"}, result); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	want := fmt.Sprintf("10.1.0.5:%d", customPort)
	if result.Status.ADBAddress != want {
		t.Errorf("expected ADBAddress=%q, got %q", want, result.Status.ADBAddress)
	}
}

// TestRedroidInstance_ReadyCondition verifies the Ready condition is True when the Pod is Running.
func TestRedroidInstance_ReadyCondition(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-cond", 0, false)

	podName := "redroid-instance-inst-cond"
	runningPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.99"},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, runningPod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-cond")
	reconcileInstance(t, r, "inst-cond")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-cond", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	found := false
	for _, c := range result.Status.Conditions {
		if c.Type == string(redroidv1alpha1.RedroidInstanceConditionReady) {
			found = true
			if c.Status != "True" {
				t.Errorf("expected Ready=True, got %v", c.Status)
			}
		}
	}
	if !found {
		t.Errorf("Ready condition not found in: %v", result.Status.Conditions)
	}
}

// TestRedroidInstance_ObservedGeneration verifies ObservedGeneration is set in status.
func TestRedroidInstance_ObservedGeneration(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-gen", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-gen")
	reconcileInstance(t, r, "inst-gen")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-gen", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.ObservedGeneration != result.Generation {
		t.Errorf("expected ObservedGeneration=%d, got %d", result.Generation, result.Status.ObservedGeneration)
	}
}

// TestRedroidInstance_Tolerations verifies tolerations are propagated to the Pod spec.
func TestRedroidInstance_Tolerations(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-tol", 0, false)
	inst.Spec.Tolerations = []corev1.Toleration{
		{Key: "gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-tol")
	reconcileInstance(t, r, "inst-tol")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-tol", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	if len(pod.Spec.Tolerations) == 0 {
		t.Error("expected tolerations to be set on Pod")
	}
	if pod.Spec.Tolerations[0].Key != "gpu" {
		t.Errorf("unexpected toleration key: %q", pod.Spec.Tolerations[0].Key)
	}
}

// TestRedroidInstance_StoppedConditionWhenSuspended verifies Scheduled=False when suspended.
func TestRedroidInstance_StoppedConditionWhenSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-susp-cond", 0, true)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-susp-cond")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-susp-cond", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
		t.Errorf("expected Stopped phase, got %v", result.Status.Phase)
	}
	for _, c := range result.Status.Conditions {
		if c.Type == string(redroidv1alpha1.RedroidInstanceConditionReady) && c.Status == "True" {
			t.Error("Ready condition should not be True when suspended")
		}
	}
}

// TestRedroidInstance_IgnoreNotFound (duplicate guard - tests the exported Reconcile via ctrl.Request)
func TestRedroidInstance_ReconcileReturnsNilForMissing(t *testing.T) {
	scheme := newTestScheme(t)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "ghost-2", Namespace: "default"},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.Requeue {
		t.Error("should not requeue for not-found resource")
	}
}

// TestRedroidInstance_BaseModeDirectMount verifies that baseMode=true mounts
// sharedDataPVC directly as /data and passes use_redroid_overlayfs=0.
func TestRedroidInstance_BaseModeDirectMount(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-base", 0, false)
	inst.Spec.BaseMode = true

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-base")
	reconcileInstance(t, r, "inst-base")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-inst-base", Namespace: "default"}, pod); err != nil {
		t.Fatalf("expected Pod: %v", err)
	}

	// Overlayfs must be disabled in base mode.
	foundOverlayfsOff := false
	for _, a := range pod.Spec.Containers[0].Args {
		if a == "androidboot.use_redroid_overlayfs=0" {
			foundOverlayfsOff = true
		}
		if a == "androidboot.use_redroid_overlayfs=1" {
			t.Error("overlayfs must be 0 in baseMode, got 1")
		}
	}
	if !foundOverlayfsOff {
		t.Errorf("expected androidboot.use_redroid_overlayfs=0 in args; got: %v",
			pod.Spec.Containers[0].Args)
	}

	// /data must be mounted from sharedDataPVC.
	foundDataMount := false
	for _, vm := range pod.Spec.Containers[0].VolumeMounts {
		if vm.MountPath == "/data" {
			foundDataMount = true
		}
		if vm.MountPath == "/data-base" || vm.MountPath == "/data-diff/0" {
			t.Errorf("base mode must not mount overlayfs paths, found: %s", vm.MountPath)
		}
	}
	if !foundDataMount {
		t.Error("expected /data mount in base mode")
	}

	// diff volume must not be present.
	for _, v := range pod.Spec.Volumes {
		if v.Name == "data-diff" {
			t.Error("base mode must not include data-diff volume")
		}
	}
}

// TestRedroidInstance_NormalModeOverlayfs verifies that baseMode=false (default)
// keeps overlayfs=1 and mounts /data-base + /data-diff.
func TestRedroidInstance_NormalModeOverlayfs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-normal", 0, false) // baseMode defaults to false

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-normal")
	reconcileInstance(t, r, "inst-normal")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-inst-normal", Namespace: "default"}, pod); err != nil {
		t.Fatalf("expected Pod: %v", err)
	}

	// Overlayfs must be enabled in normal mode.
	foundOverlayfsOn := false
	for _, a := range pod.Spec.Containers[0].Args {
		if a == "androidboot.use_redroid_overlayfs=1" {
			foundOverlayfsOn = true
		}
	}
	if !foundOverlayfsOn {
		t.Errorf("expected androidboot.use_redroid_overlayfs=1, got: %v", pod.Spec.Containers[0].Args)
	}

	// /data-base and data-diff must be present.
	mounts := map[string]bool{}
	for _, vm := range pod.Spec.Containers[0].VolumeMounts {
		mounts[vm.MountPath] = true
	}
	if !mounts["/data-base"] {
		t.Error("expected /data-base mount in normal mode")
	}
	if !mounts["/data-diff/0"] {
		t.Error("expected /data-diff/0 mount in normal mode")
	}
}
