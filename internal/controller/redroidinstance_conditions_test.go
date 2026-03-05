package controller_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// findInstanceCondition returns the condition of the given type, or nil if absent.
func findInstanceCondition(conds []metav1.Condition, condType string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == condType {
			return &conds[i]
		}
	}
	return nil
}

// getInstanceAfterReconcile runs two reconcile passes (first adds finalizer, second does work)
// and returns the refreshed RedroidInstance.
func getInstanceAfterReconcile(t *testing.T, r *controller.RedroidInstanceReconciler, name string) *redroidv1alpha1.RedroidInstance {
	t.Helper()
	reconcileInstance(t, r, name) // adds finalizer
	reconcileInstance(t, r, name) // actual work
	result := &redroidv1alpha1.RedroidInstance{}
	if err := r.Get(context.Background(),
		types.NamespacedName{Name: name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get instance %q: %v", name, err)
	}
	return result
}

// ── Ready condition ──────────────────────────────────────────────────────────

// TestInstanceCondition_Ready_Running verifies that a Running pod with an IP sets
// Ready=True with Reason=Running and the ADB address in the message.
func TestInstanceCondition_Ready_Running(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-run", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.1.2.3"},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Ready=True, got %v", cond.Status)
	}
	if cond.Reason != "Running" {
		t.Errorf("expected Reason=Running, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, "10.1.2.3:5555") {
		t.Errorf("expected ADB address in message, got %q", cond.Message)
	}
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_PodRunningNoADB verifies Reason=PodRunningNoADB when
// the pod is Running but has no IP yet.
func TestInstanceCondition_Ready_PodRunningNoADB(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-noadb", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: ""},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Ready=False, got %v", cond.Status)
	}
	if cond.Reason != "PodRunningNoADB" {
		t.Errorf("expected Reason=PodRunningNoADB, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_Pending verifies Reason=Pending when pod was just created.
func TestInstanceCondition_Ready_Pending(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-pend", 0, false)
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Ready=False, got %v", cond.Status)
	}
	if cond.Reason != "Pending" {
		t.Errorf("expected Reason=Pending, got %q", cond.Reason)
	}
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_Failed_WithDetails verifies that when a Pod has
// failed containers the Ready condition message includes the container name,
// exit code, and reason extracted from the termination status.
func TestInstanceCondition_Ready_Failed_WithDetails(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-fail-detail", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	exitCode := int32(137)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "redroid",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: exitCode,
							Reason:   "OOMKilled",
						},
					},
				},
			},
		},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Ready=False, got %v", cond.Status)
	}
	if cond.Reason != "PodFailed" {
		t.Errorf("expected Reason=PodFailed, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, "137") {
		t.Errorf("expected exit code 137 in message, got %q", cond.Message)
	}
	if !strings.Contains(cond.Message, "OOMKilled") {
		t.Errorf("expected reason OOMKilled in message, got %q", cond.Message)
	}
	if !strings.Contains(cond.Message, `"redroid"`) {
		t.Errorf("expected container name in message, got %q", cond.Message)
	}
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_Failed_NoDetails verifies the fallback message when
// the pod phase is Failed but no container termination info is available.
func TestInstanceCondition_Ready_Failed_NoDetails(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-fail-nodetail", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Reason != "PodFailed" {
		t.Errorf("expected Reason=PodFailed, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in fallback message, got %q", cond.Message)
	}
	// Fallback message must say "Inspect" rather than showing empty details.
	if !strings.Contains(strings.ToLower(cond.Message), "inspect") {
		t.Errorf("expected 'inspect' in fallback message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_Stopped_SpecSuspend verifies the Stopped message
// mentions spec.suspend when spec.suspend=true.
func TestInstanceCondition_Ready_Stopped_SpecSuspend(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-stop-spec", 0, true) // spec.suspend=true
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Reason != "Stopped" {
		t.Errorf("expected Reason=Stopped, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, "spec.suspend") {
		t.Errorf("expected 'spec.suspend' in message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Ready_Stopped_StatusSuspended verifies the Stopped message
// includes the actor name when status.suspended is set.
func TestInstanceCondition_Ready_Stopped_StatusSuspended(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-stop-actor", 0, false) // spec.suspend=false
	// Pre-set status.suspended with a known actor.
	inst.Status = redroidv1alpha1.RedroidInstanceStatus{
		Suspended: &redroidv1alpha1.SuspendedStatus{
			Actor:  "task/my-task",
			Reason: "reserved for one-shot task",
		},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Ready")
	if cond == nil {
		t.Fatal("Ready condition not found")
	}
	if cond.Reason != "Stopped" {
		t.Errorf("expected Reason=Stopped, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, "task/my-task") {
		t.Errorf("expected actor 'task/my-task' in message, got %q", cond.Message)
	}
}

// ── Scheduled condition ──────────────────────────────────────────────────────

// TestInstanceCondition_Scheduled_PodCreated verifies Scheduled=True with the
// pod name in the message when a pod exists and instance is not failed/stopped.
func TestInstanceCondition_Scheduled_PodCreated(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-sched-ok", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.5"},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
	if cond == nil {
		t.Fatal("Scheduled condition not found")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Scheduled=True, got %v", cond.Status)
	}
	if cond.Reason != "PodCreated" {
		t.Errorf("expected Reason=PodCreated, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, podName) {
		t.Errorf("expected pod name in Scheduled message, got %q", cond.Message)
	}
}

// TestInstanceCondition_Scheduled_FailedPod verifies Scheduled=True with a
// "created but has since failed" message when phase=Failed.
func TestInstanceCondition_Scheduled_FailedPod(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-sched-fail", 0, false)
	podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed},
	}
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst, pod).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
	if cond == nil {
		t.Fatal("Scheduled condition not found")
	}
	// Pod was created (Scheduled=True) but is now failed.
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Scheduled=True for failed pod, got %v", cond.Status)
	}
	if cond.Reason != "PodCreated" {
		t.Errorf("expected Reason=PodCreated, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, "failed") {
		t.Errorf("expected 'failed' in Scheduled message for failed pod, got %q", cond.Message)
	}
}

// TestInstanceCondition_Scheduled_Stopped verifies Scheduled=False with
// Reason=Stopped when instance is stopped (spec.suspend=true).
func TestInstanceCondition_Scheduled_Stopped(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("cond-sched-stop", 0, true)
	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
	result := getInstanceAfterReconcile(t, r, inst.Name)

	cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
	if cond == nil {
		t.Fatal("Scheduled condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Scheduled=False when stopped, got %v", cond.Status)
	}
	if cond.Reason != "Stopped" {
		t.Errorf("expected Reason=Stopped, got %q", cond.Reason)
	}
}
