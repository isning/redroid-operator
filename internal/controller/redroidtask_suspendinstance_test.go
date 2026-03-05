package controller_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// makeStoppedInstance creates an instance whose status already reports Stopped
// (simulating the instance controller having reacted to a suspended).
func makeStoppedInstance(name string, index int) *redroidv1alpha1.RedroidInstance {
	return &redroidv1alpha1.RedroidInstance{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: redroidv1alpha1.RedroidInstanceSpec{
			Index:         index,
			Image:         "redroid/redroid:16.0.0-latest",
			SharedDataPVC: "redroid-data-base-pvc",
			DiffDataPVC:   "redroid-data-diff-pvc",
			GPUMode:       "host",
		},
		Status: redroidv1alpha1.RedroidInstanceStatus{
			Phase: redroidv1alpha1.RedroidInstanceStopped,
		},
	}
}

// makeTaskSuspendInstance creates a one-shot task with SuspendInstance=true.
func makeTaskSuspendInstance(name string, instances []string) *redroidv1alpha1.RedroidTask {
	task := makeTask(name, instances, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	task.Spec.SuspendInstance = true
	return task
}

var _ = Describe("RedroidTask SuspendInstance", func() {
	var (
		scheme = newTestScheme()
	)

	It("sets TempSuspend on instance before creating Job and requeues", func() {
		inst := makeRunningInstance("maa-sus", 0, "10.0.0.1:5555")
		task := makeTaskSuspendInstance("task-sus", []string{"maa-sus"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}

		// 1st reconcile: adds finalizer.
		reconcileTask(r, "task-sus")
		// 2nd reconcile: should set suspended on instance and requeue.
		res := reconcileTask(r, "task-sus")

		Expect(res.RequeueAfter).NotTo(Equal(0), "expected RequeueAfter > 0 while waiting for instance to stop")

		// Instance should now have suspended set.
		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-sus", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Suspended).NotTo(BeNil(), "expected suspended to be set on instance")
		Expect(updatedInst.Status.Suspended.Actor).To(Equal("task/task-sus"))

		// Job must NOT have been created yet.
		jobList := &batchv1.JobList{}
		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "expected 0 jobs (instance not stopped yet)")
	})

	It("keeps requeueing without creating Job when instance is suspended but not Stopped", func() {
		// Instance is Running but has suspended already set.
		inst := makeRunningInstance("maa-waits", 0, "10.0.0.1:5555")
		inst.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Actor:  "task/task-waits",
			Reason: "reserved for one-shot task task-waits",
		}
		task := makeTaskSuspendInstance("task-waits", []string{"maa-waits"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
		reconcileTask(r, "task-waits") // add finalizer

		res := reconcileTask(r, "task-waits") // should requeue, no job
		Expect(res.RequeueAfter).NotTo(Equal(0), "expected RequeueAfter > 0 while instance is still Running")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "job must not be created before instance is Stopped")
	})

	It("creates Job once instance is suspended and Stopped", func() {
		inst := makeStoppedInstance("maa-ready", 0)
		inst.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Actor:  "task/task-ready",
			Reason: "reserved for one-shot task task-ready",
		}
		task := makeTaskSuspendInstance("task-ready", []string{"maa-ready"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
		reconcileTask(r, "task-ready") // add finalizer
		reconcileTask(r, "task-ready") // create Job

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(HaveLen(1), "expected 1 job after instance stopped")
	})

	It("clears status.suspended after Job is done", func() {
		inst := makeStoppedInstance("maa-done", 0)
		inst.Status.Suspended = &redroidv1alpha1.SuspendedStatus{
			Actor:  "task/task-done",
			Reason: "reserved for one-shot task task-done",
		}
		task := makeTaskSuspendInstance("task-done", []string{"maa-done"})

		// Create a finished Job manually.
		trueVal := true
		finishedJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "task-done-maa-done",
				Namespace: "default",
				Labels:    map[string]string{"app.kubernetes.io/managed-by": "redroid-operator"},
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
				},
				CompletionTime: &metav1.Time{},
			},
		}
		// Set controller reference equivalent field for ownership.
		finishedJob.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "redroid.isning.moe/v1alpha1",
				Kind:       "RedroidTask",
				Name:       task.Name,
				Controller: &trueVal,
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task, finishedJob).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
		reconcileTask(r, "task-done") // add finalizer
		reconcileTask(r, "task-done") // detect finished job → clear suspended

		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-done", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Suspended).To(BeNil(), "expected suspended to be cleared after Job completion")
	})
})
