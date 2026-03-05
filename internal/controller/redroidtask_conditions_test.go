package controller_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// findTaskCondition returns the condition of the given type from the task status,
// or nil if absent.
func findTaskCondition(conds []metav1.Condition, condType string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == condType {
			return &conds[i]
		}
	}
	return nil
}

var _ = Describe("RedroidTask Conditions", func() {
	var (
		scheme = newTestScheme()
	)

	Describe("Active condition", func() {
		It("sets Active=False with Reason=NoActiveJobs when all jobs are finished", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("cond-active-none", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name) // adds finalizer
			reconcileTask(r, task.Name) // creates Job

			// Mark the job as Complete so it is no longer tracked as active.
			jobList := &batchv1.JobList{}
			err := fc.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty(), "expected a job to be created")

			job := jobList.Items[0].DeepCopy()
			job.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}
			err = fc.Status().Update(context.Background(), job)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, task.Name) // patchTaskStatus with no active jobs

			result := &redroidv1alpha1.RedroidTask{}
			err = fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Active")
			Expect(cond).NotTo(BeNil(), "Active condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("NoActiveJobs"))
		})

		It("sets Active=True with Reason=JobsActive and job names listed when a job is running", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("cond-active-run", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name) // job created, not yet finished

			result := &redroidv1alpha1.RedroidTask{}
			err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Active")
			Expect(cond).NotTo(BeNil(), "Active condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("JobsActive"))

			// Job name should be embedded in the message.
			expectedJobName := task.Name + "-maa-0"
			Expect(cond.Message).To(ContainSubstring(expectedJobName))
		})
	})

	Describe("Failed condition", func() {
		It("sets Failed=True when a job reports Failed condition, and embeds reason in message", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("cond-fail-job", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name) // job created

			// Simulate job failure with a reason.
			jobList := &batchv1.JobList{}
			err := fc.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty(), "expected a job to be created")

			job := jobList.Items[0].DeepCopy()
			job.Status.Conditions = []batchv1.JobCondition{
				{
					Type:    batchv1.JobFailed,
					Status:  corev1.ConditionTrue,
					Reason:  "BackoffLimitExceeded",
					Message: "Job has reached the specified backoff limit",
				},
			}
			err = fc.Status().Update(context.Background(), job)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, task.Name)

			result := &redroidv1alpha1.RedroidTask{}
			err = fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Failed")
			Expect(cond).NotTo(BeNil(), "Failed condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("JobsFailed"))
			Expect(cond.Message).To(ContainSubstring(job.Name))
			Expect(cond.Message).To(ContainSubstring("BackoffLimitExceeded"))
			Expect(cond.Message).To(ContainSubstring("backoff limit"))
		})

		It("sets Failed=False when all jobs succeed", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("cond-fail-none", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name)

			// Mark the job as Complete (success).
			jobList := &batchv1.JobList{}
			err := fc.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty())

			job := jobList.Items[0].DeepCopy()
			job.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}
			err = fc.Status().Update(context.Background(), job)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, task.Name)

			result := &redroidv1alpha1.RedroidTask{}
			err = fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Failed")
			Expect(cond).NotTo(BeNil(), "Failed condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("NoFailedJobs"))
		})
	})

	Describe("Complete condition", func() {
		It("sets Complete=True with Reason=AllJobsSucceeded when all jobs complete", func() {
			inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
			task := makeTask("cond-complete-all", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst0, inst1, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name) // 2 jobs created

			// Mark both jobs as Complete.
			jobList := &batchv1.JobList{}
			err := fc.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).To(HaveLen(2))

			for i := range jobList.Items {
				j := jobList.Items[i].DeepCopy()
				j.Status.Conditions = []batchv1.JobCondition{
					{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
				}
				err = fc.Status().Update(context.Background(), j)
				Expect(err).NotTo(HaveOccurred())
			}

			reconcileTask(r, task.Name)

			result := &redroidv1alpha1.RedroidTask{}
			err = fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Complete")
			Expect(cond).NotTo(BeNil(), "Complete condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("AllJobsSucceeded"))
			Expect(cond.Message).To(ContainSubstring("2"))
		})

		It("sets Complete=False and shows fractional progress when some jobs complete", func() {
			inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
			task := makeTask("cond-complete-partial", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
				WithObjects(inst0, inst1, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name) // 2 jobs created

			// Mark only first job as complete; second stays running.
			jobList := &batchv1.JobList{}
			err := fc.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).To(HaveLen(2))

			first := jobList.Items[0].DeepCopy()
			first.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			}
			err = fc.Status().Update(context.Background(), first)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, task.Name)

			result := &redroidv1alpha1.RedroidTask{}
			err = fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			cond := findTaskCondition(result.Status.Conditions, "Complete")
			Expect(cond).NotTo(BeNil(), "Complete condition not found")
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			// Message should express fractional progress.
			Expect(cond.Message).To(ContainSubstring("1"))
			Expect(cond.Message).To(ContainSubstring("2"))
		})
	})

	Describe("CronJob (scheduled) task conditions", func() {
		It("provides stable 'Scheduled' reason for Complete and Failed conditions in CronJobs", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("cond-cron-sched", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fc := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()
			r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileTask(r, task.Name)
			reconcileTask(r, task.Name)

			result := &redroidv1alpha1.RedroidTask{}
			err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			for _, condType := range []string{"Complete", "Failed"} {
				cond := findTaskCondition(result.Status.Conditions, condType)
				Expect(cond).NotTo(BeNil(), "%s condition not found", condType)
				Expect(cond.Reason).To(Equal("Scheduled"))
				Expect(strings.ToLower(cond.Message)).To(ContainSubstring("cronjob"))
			}
		})
	})
})
