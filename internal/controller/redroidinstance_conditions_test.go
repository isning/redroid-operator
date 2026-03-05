package controller_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
func getInstanceAfterReconcile(r *controller.RedroidInstanceReconciler, name string) *redroidv1alpha1.RedroidInstance {
	reconcileInstance(r, name) // adds finalizer
	reconcileInstance(r, name) // actual work
	result := &redroidv1alpha1.RedroidInstance{}
	err := r.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "default"}, result)
	Expect(err).NotTo(HaveOccurred(), "get instance %q", name)
	return result
}

var _ = Describe("RedroidInstance Conditions", func() {
	var (
		scheme = newTestScheme()
	)

	Describe("Ready condition", func() {
		It("sets Ready=True with Reason=Running and ADB address when Pod is Running", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("Running"))
			Expect(cond.Message).To(ContainSubstring("10.1.2.3:5555"))
			Expect(cond.Message).To(ContainSubstring(podName))
		})

		It("sets Reason=PodRunningNoADB when Pod is Running without IP", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("PodRunningNoADB"))
			Expect(cond.Message).To(ContainSubstring(podName))
		})

		It("sets Reason=Pending when pod is just created", func() {
			inst := makeInstance("cond-pend", 0, false)
			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("Pending"))

			podName := fmt.Sprintf("redroid-instance-%s", inst.Name)
			Expect(cond.Message).To(ContainSubstring(podName))
		})

		It("extracts container termination details on failed pod", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("PodFailed"))
			Expect(cond.Message).To(ContainSubstring("137"))
			Expect(cond.Message).To(ContainSubstring("OOMKilled"))
			Expect(cond.Message).To(ContainSubstring(`"redroid"`))
			Expect(cond.Message).To(ContainSubstring(podName))
		})

		It("provides a fallback message on failed pod without termination info", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Reason).To(Equal("PodFailed"))
			Expect(cond.Message).To(ContainSubstring(podName))
			Expect(strings.ToLower(cond.Message)).To(ContainSubstring("inspect"))
		})

		It("mentions spec.suspend when spec.suspend=true", func() {
			inst := makeInstance("cond-stop-spec", 0, true) // spec.suspend=true
			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Reason).To(Equal("Stopped"))
			Expect(cond.Message).To(ContainSubstring("spec.suspend"))
		})

		It("mentions actor when status.suspended is set", func() {
			inst := makeInstance("cond-stop-actor", 0, false) // spec.suspend=false
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil(), "Ready condition not found")
			Expect(cond.Reason).To(Equal("Stopped"))
			Expect(cond.Message).To(ContainSubstring("task/my-task"))
		})
	})

	Describe("Scheduled condition", func() {
		It("sets Scheduled=True and includes pod name when pod exists", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
			Expect(cond).NotTo(BeNil(), "Scheduled condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("PodCreated"))
			Expect(cond.Message).To(ContainSubstring(podName))
		})

		It("sets Scheduled=True but mentions failures if phase=Failed", func() {
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
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
			Expect(cond).NotTo(BeNil(), "Scheduled condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("PodCreated"))
			Expect(cond.Message).To(ContainSubstring("failed"))
		})

		It("sets Scheduled=False with Reason=Stopped when instance is suspended", func() {
			inst := makeInstance("cond-sched-stop", 0, true)
			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r := &controller.RedroidInstanceReconciler{Client: fc, Scheme: scheme}
			result := getInstanceAfterReconcile(r, inst.Name)

			cond := findInstanceCondition(result.Status.Conditions, "Scheduled")
			Expect(cond).NotTo(BeNil(), "Scheduled condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("Stopped"))
		})
	})
})
