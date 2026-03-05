package controller_test

// Tests for status.suspended — the programmatic suspend mechanism
// that avoids Flux/GitOps config drift.

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

var _ = Describe("RedroidInstance TempSuspend", func() {
	var (
		scheme = newTestScheme()
	)

	It("stops the Pod and changes phase to Stopped when status.suspended is set (with spec.suspend=false)", func() {
		inst := makeInstance("inst-tmpsus", 0, false) // spec.suspend = false

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

		// 1st reconcile: add finalizer.
		reconcileInstance(r, "inst-tmpsus")
		// 2nd reconcile: create Pod.
		reconcileInstance(r, "inst-tmpsus")

		// Confirm Pod is created.
		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-tmpsus", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "expected Pod to exist")

		// Set status.suspended without touching spec.suspend.
		fresh := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-tmpsus", Namespace: "default"}, fresh)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Reason: "maintenance",
			Actor:  "manual",
		}
		err = fakeClient.Status().Update(context.Background(), fresh)
		Expect(err).NotTo(HaveOccurred(), "set suspended")

		// Reconcile — controller should delete Pod and set phase=Stopped.
		reconcileInstance(r, "inst-tmpsus")

		result := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-tmpsus", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get after reconcile")

		Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceStopped))
		Expect(result.Spec.Suspend).To(BeFalse(), "spec.suspend must not be modified by suspended logic")

		// Pod should be gone.
		deletedPod := &corev1.Pod{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-tmpsus", Namespace: "default"}, deletedPod)
		Expect(err).To(HaveOccurred(), "Pod should have been deleted when suspended is set")
	})

	It("auto-clears expired status.suspended.until and does not delete Pod", func() {
		inst := makeInstance("inst-expiry", 0, false)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-expiry") // add finalizer
		reconcileInstance(r, "inst-expiry") // create Pod

		// Set an already-expired suspended.
		fresh := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-expiry", Namespace: "default"}, fresh)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		past := metav1.NewTime(time.Now().Add(-1 * time.Hour))
		fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Reason: "expired maintenance window",
			Actor:  "manual",
			Until:  &past,
		}
		err = fakeClient.Status().Update(context.Background(), fresh)
		Expect(err).NotTo(HaveOccurred(), "set suspended")

		// Reconcile — expired suspend should be auto-cleared; Pod should NOT be deleted.
		reconcileInstance(r, "inst-expiry")

		result := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-expiry", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get after reconcile")

		Expect(result.Status.Suspended).To(BeNil(), "expired suspended should have been cleared")

		// Pod should still exist.
		liveP := &corev1.Pod{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-expiry", Namespace: "default"}, liveP)
		Expect(err).NotTo(HaveOccurred(), "Pod should still exist after expired suspend")
	})

	It("stops the Pod when both spec.suspend and status.suspended are set", func() {
		inst := makeInstance("inst-both-sus", 0, true) // spec.suspend = true

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

		// Also set suspended.
		fresh := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-both-sus", Namespace: "default"}, fresh)
		// FakeClient Get fails if object wasn't created via Reconcile first for some reason?
		// Actually fake client has it from WithObjects.
		Expect(err).NotTo(HaveOccurred(), "get instance")

		fresh.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Reason: "task/maa needs exclusive access",
			Actor:  "task/maa",
		}
		err = fakeClient.Status().Update(context.Background(), fresh)
		Expect(err).NotTo(HaveOccurred(), "set suspended")

		reconcileInstance(r, "inst-both-sus")

		result := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-both-sus", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get after reconcile")

		Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceStopped))
	})
})
