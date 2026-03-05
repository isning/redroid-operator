package controller_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

var _ = Describe("RedroidTask Features", func() {
	var (
		scheme = newTestScheme()
	)

	It("Applies custom BackoffLimit to the Job", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-bl", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
		bl := int32(5)
		task.Spec.BackoffLimit = &bl

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-bl")
		reconcileTask(r, "task-bl")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

		Expect(jobList.Items[0].Spec.BackoffLimit).NotTo(BeNil())
		Expect(*jobList.Items[0].Spec.BackoffLimit).To(Equal(int32(5)))
	})

	It("Applies TTLSecondsAfterFinished to one-shot Jobs", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-ttl", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
		ttl := int32(300)
		task.Spec.TTLSecondsAfterFinished = &ttl

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-ttl")
		reconcileTask(r, "task-ttl")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

		Expect(jobList.Items[0].Spec.TTLSecondsAfterFinished).NotTo(BeNil())
		Expect(*jobList.Items[0].Spec.TTLSecondsAfterFinished).To(Equal(int32(300)))
	})

	It("Applies Timezone to CronJob", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-tz", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
		task.Spec.Timezone = "Asia/Shanghai"

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-tz")
		reconcileTask(r, "task-tz")

		cronList := &batchv1.CronJobList{}
		err := fakeClient.List(context.Background(), cronList)
		Expect(err).NotTo(HaveOccurred(), "list cronjobs")
		Expect(cronList.Items).NotTo(BeEmpty(), "expected CronJob")

		Expect(cronList.Items[0].Spec.TimeZone).NotTo(BeNil())
		Expect(*cronList.Items[0].Spec.TimeZone).To(Equal("Asia/Shanghai"))
	})

	It("Sets ObservedGeneration in task status", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-gen", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-gen")
		reconcileTask(r, "task-gen")

		result := &redroidv1alpha1.RedroidTask{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-gen", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get task")
		Expect(result.Status.ObservedGeneration).To(Equal(result.Generation))
	})

	It("Applies WorkingDir to integration containers", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		integration := redroidv1alpha1.IntegrationSpec{
			Name:       "tool",
			Image:      "tool:latest",
			WorkingDir: "/workspace",
		}
		task := makeTask("task-wd", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{integration})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-wd")
		reconcileTask(r, "task-wd")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

		containers := jobList.Items[0].Spec.Template.Spec.Containers
		Expect(containers).NotTo(BeEmpty(), "no containers")
		Expect(containers[0].WorkingDir).To(Equal("/workspace"))
	})

	It("Applies StartingDeadlineSeconds to CronJob", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-sds", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
		sds := int64(180)
		task.Spec.StartingDeadlineSeconds = &sds

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-sds")
		reconcileTask(r, "task-sds")

		cronList := &batchv1.CronJobList{}
		err := fakeClient.List(context.Background(), cronList)
		Expect(err).NotTo(HaveOccurred(), "list cronjobs")
		Expect(cronList.Items).NotTo(BeEmpty(), "expected CronJob")

		Expect(cronList.Items[0].Spec.StartingDeadlineSeconds).NotTo(BeNil())
		Expect(*cronList.Items[0].Spec.StartingDeadlineSeconds).To(Equal(int64(180)))
	})

	It("Propagates ImagePullSecrets to the Pod spec", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		task := makeTask("task-ips", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
		task.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "my-registry-secret"}}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-ips")
		reconcileTask(r, "task-ips")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

		pullSecrets := jobList.Items[0].Spec.Template.Spec.ImagePullSecrets
		found := false
		for _, s := range pullSecrets {
			if s.Name == "my-registry-secret" {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "expected ImagePullSecret 'my-registry-secret' in pod spec, got: %v", pullSecrets)
	})

	It("Propagates ServiceAccountName to the Job's PodSpec", func() {
		inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
		integ := basicIntegration()
		task := makeTask("task-sa", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{integ})
		task.Spec.ServiceAccountName = "my-task-sa"

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-sa")
		reconcileTask(r, "task-sa")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).NotTo(BeEmpty(), "no jobs created")

		got := jobList.Items[0].Spec.Template.Spec.ServiceAccountName
		Expect(got).To(Equal("my-task-sa"))
	})

	It("Completes wakeInstance=true lifecycle (base-init scenario)", func() {
		// suspended instance → task sets status.woken → instance becomes Running
		// → Job created → Job finishes → status.woken cleared
		inst := makeSuspendedSpecInstance("maa-base", 0)
		task := makeTaskWakeInstance("base-init-lifecycle", []string{"maa-base"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

		// Pass 1: add finalizer.
		reconcileTask(r, "base-init-lifecycle")

		// Pass 2: controller sets status.woken and requeues without creating a Job.
		res, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "base-init-lifecycle", Namespace: "default"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).NotTo(Equal(0), "expected requeue while waiting for instance to start")

		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-base", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Woken).NotTo(BeNil(), "expected status.woken to be set on instance")

		jobList := &batchv1.JobList{}
		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "expected 0 jobs before instance is Running")

		// Simulate instance controller: instance transitions to Running.
		updatedInst.Status.Phase = redroidv1alpha1.RedroidInstanceRunning
		updatedInst.Status.ADBAddress = "10.0.0.5:5555"
		err = fakeClient.Status().Update(context.Background(), updatedInst)
		Expect(err).NotTo(HaveOccurred(), "update instance phase to Running")

		// Pass 3: Job is created now that instance is Running.
		reconcileTask(r, "base-init-lifecycle")

		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(HaveLen(1), "expected 1 Job after instance is Running")

		// Simulate Job completion.
		job := &jobList.Items[0]
		now := metav1.Now()
		job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		})
		job.Status.CompletionTime = &now
		err = fakeClient.Status().Update(context.Background(), job)
		Expect(err).NotTo(HaveOccurred(), "mark job complete")

		// Pass 4: controller detects finished Job and clears status.woken.
		reconcileTask(r, "base-init-lifecycle")

		finalInst := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-base", Namespace: "default"}, finalInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(finalInst.Status.Woken).To(BeNil(), "expected status.woken to be cleared after Job completion")
	})

	It("Completes suspendInstance=true lifecycle (base-update scenario)", func() {
		// running instance → task sets status.suspended → instance becomes Stopped
		// → Job created → Job finishes → status.suspended cleared
		inst := makeRunningInstance("maa-base-su", 0, "10.0.0.6:5555")
		task := makeTaskSuspendInstance("base-update-lifecycle", []string{"maa-base-su"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

		// Pass 1: add finalizer.
		reconcileTask(r, "base-update-lifecycle")

		// Pass 2: controller sets status.suspended and requeues without creating a Job.
		res, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "base-update-lifecycle", Namespace: "default"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.RequeueAfter).NotTo(Equal(0), "expected requeue while waiting for instance to stop")

		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-base-su", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Suspended).NotTo(BeNil(), "expected status.suspended to be set on instance")

		jobList := &batchv1.JobList{}
		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "expected 0 jobs before instance is Stopped")

		// Simulate instance controller: instance transitions to Stopped.
		updatedInst.Status.Phase = redroidv1alpha1.RedroidInstanceStopped
		err = fakeClient.Status().Update(context.Background(), updatedInst)
		Expect(err).NotTo(HaveOccurred(), "update instance phase to Stopped")

		// Pass 3: Job is created now that instance is Stopped.
		reconcileTask(r, "base-update-lifecycle")

		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(HaveLen(1), "expected 1 Job after instance is Stopped")

		// Simulate Job completion.
		job := &jobList.Items[0]
		now := metav1.Now()
		job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		})
		job.Status.CompletionTime = &now
		err = fakeClient.Status().Update(context.Background(), job)
		Expect(err).NotTo(HaveOccurred(), "mark job complete")

		// Pass 4: controller detects finished Job and clears status.suspended.
		reconcileTask(r, "base-update-lifecycle")

		finalInst := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-base-su", Namespace: "default"}, finalInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(finalInst.Status.Suspended).To(BeNil(), "expected status.suspended to be cleared after Job completion")
	})
})
