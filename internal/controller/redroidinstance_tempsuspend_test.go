package controller_test

// Tests for status.suspended — the programmatic suspend mechanism
// that avoids Flux/GitOps config drift.

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// TestRedroidInstance_SuspendedStopsPod verifies that setting
// status.suspended (with spec.suspend=false) causes the running Pod to
// be deleted and the phase changed to Stopped.
func TestRedroidInstance_SuspendedStopsPod(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-tmpsus", 0, false) // spec.suspend = false

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}

	// 1st reconcile: add finalizer.
	reconcileInstance(t, r, "inst-tmpsus")
	// 2nd reconcile: create Pod.
	reconcileInstance(t, r, "inst-tmpsus")

	// Confirm Pod is created.
	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-inst-tmpsus", Namespace: "default"}, pod); err != nil {
		t.Fatalf("expected Pod to exist: %v", err)
	}

	// Set status.suspended without touching spec.suspend.
	fresh := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-tmpsus", Namespace: "default"}, fresh); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
		Reason: "maintenance",
		Actor:  "manual",
	}
	if err := fakeClient.Status().Update(context.Background(), fresh); err != nil {
		t.Fatalf("set suspended: %v", err)
	}

	// Reconcile — controller should delete Pod and set phase=Stopped.
	reconcileInstance(t, r, "inst-tmpsus")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-tmpsus", Namespace: "default"}, result); err != nil {
		t.Fatalf("get after reconcile: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
		t.Errorf("expected Stopped phase, got %v", result.Status.Phase)
	}
	// spec.suspend should remain false — that is the whole point.
	if result.Spec.Suspend {
		t.Error("spec.suspend must not be modified by suspended logic")
	}
	// Pod should be gone.
	deletedPod := &corev1.Pod{}
	err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-inst-tmpsus", Namespace: "default"}, deletedPod)
	if err == nil {
		t.Error("Pod should have been deleted when suspended is set")
	}
}

// TestRedroidInstance_SuspendedExpiry verifies that once
// status.suspended.until is in the past the field is auto-cleared and
// the instance Pod is not deleted.
func TestRedroidInstance_SuspendedExpiry(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-expiry", 0, false)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-expiry") // add finalizer
	reconcileInstance(t, r, "inst-expiry") // create Pod

	// Set an already-expired suspended.
	fresh := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-expiry", Namespace: "default"}, fresh); err != nil {
		t.Fatalf("get: %v", err)
	}
	past := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
		Reason: "expired maintenance window",
		Actor:  "manual",
		Until:  &past,
	}
	if err := fakeClient.Status().Update(context.Background(), fresh); err != nil {
		t.Fatalf("set suspended: %v", err)
	}

	// Reconcile — expired suspend should be auto-cleared; Pod should NOT be deleted.
	reconcileInstance(t, r, "inst-expiry")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-expiry", Namespace: "default"}, result); err != nil {
		t.Fatalf("get after reconcile: %v", err)
	}
	if result.Status.Suspended != nil {
		t.Errorf("expired suspended should have been cleared, got %+v", result.Status.Suspended)
	}
	// Pod should still exist.
	liveP := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "redroid-instance-inst-expiry", Namespace: "default"}, liveP); err != nil {
		t.Errorf("Pod should still exist after expired suspend: %v", err)
	}
}

// TestRedroidInstance_SuspendedAndSpecSuspend verifies that when both
// spec.suspend and status.suspended are set the Pod is stopped.
func TestRedroidInstance_SuspendedAndSpecSuspend(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-both-sus", 0, true) // spec.suspend = true

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}

	// Also set suspended.
	fresh := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-both-sus", Namespace: "default"}, fresh); err != nil {
		t.Fatalf("get: %v", err)
	}
	fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
		Reason: "task/maa needs exclusive access",
		Actor:  "task/maa",
	}
	if err := fakeClient.Status().Update(context.Background(), fresh); err != nil {
		t.Fatalf("set suspended: %v", err)
	}

	reconcileInstance(t, r, "inst-both-sus")

	result := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "inst-both-sus", Namespace: "default"}, result); err != nil {
		t.Fatalf("get after reconcile: %v", err)
	}
	if result.Status.Phase != redroidv1alpha1.RedroidInstanceStopped {
		t.Errorf("expected Stopped, got %v", result.Status.Phase)
	}
}
