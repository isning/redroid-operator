package controller_test

// Tests for task.Spec.WakeInstance — the mechanism where a one-shot Task
// temporarily starts the referenced RedroidInstance before running and
// releases the wake-override afterwards so the instance returns to its
// normal spec.suspend state.

import (
	"context"

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

// makeSuspendedSpecInstance creates an instance with spec.suspend=true,
// already in phase=Stopped (the instance controller has reacted).
func makeSuspendedSpecInstance(name string, index int) *redroidv1alpha1.RedroidInstance {
	return &redroidv1alpha1.RedroidInstance{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: redroidv1alpha1.RedroidInstanceSpec{
			Index:         index,
			Image:         "redroid/redroid:16.0.0-latest",
			SharedDataPVC: "redroid-data-base-pvc",
			DiffDataPVC:   "redroid-data-diff-pvc",
			GPUMode:       "host",
			Suspend:       true,
		},
		Status: redroidv1alpha1.RedroidInstanceStatus{
			Phase: redroidv1alpha1.RedroidInstanceStopped,
		},
	}
}

// makeTaskWakeInstance creates a one-shot task with WakeInstance=true.
func makeTaskWakeInstance(name string, instances []string) *redroidv1alpha1.RedroidTask {
	task := makeTask(name, instances, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	task.Spec.WakeInstance = true
	return task
}

var _ = Describe("RedroidTask WakeInstance", func() {
	var (
		scheme = newTestScheme()
	)

	It("sets status.woken on instance before creating Job and requeues", func() {
		inst := makeSuspendedSpecInstance("maa-wk", 0)
		task := makeTaskWakeInstance("task-wk", []string{"maa-wk"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

		// 1st reconcile: adds finalizer.
		reconcileTask(r, "task-wk")
		// 2nd reconcile: should set woken on instance and requeue.
		res := reconcileTask(r, "task-wk")

		Expect(res.RequeueAfter).NotTo(Equal(0), "expected RequeueAfter > 0 while waiting for instance to start")

		// Instance should now have woken set.
		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-wk", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Woken).NotTo(BeNil(), "expected woken to be set on instance")
		Expect(updatedInst.Status.Woken.Actor).To(Equal("task/task-wk"))

		// Job must NOT have been created yet.
		jobList := &batchv1.JobList{}
		err = fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "expected 0 jobs (instance not running yet)")
	})

	It("keeps requeueing without creating Job when woken is set but not Running", func() {
		// Instance is Stopped but has woken already set.
		inst := makeSuspendedSpecInstance("maa-wkwait", 0)
		inst.Status.Woken = &redroidv1alpha1.WokenStatus{
			Actor:  "task/task-wkwait",
			Reason: "on-demand wake for one-shot task task-wkwait",
		}
		task := makeTaskWakeInstance("task-wkwait", []string{"maa-wkwait"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-wkwait") // add finalizer

		res := reconcileTask(r, "task-wkwait") // should requeue, no job
		Expect(res.RequeueAfter).NotTo(Equal(0), "expected RequeueAfter > 0 while instance is still Stopped")

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(BeEmpty(), "job must not be created before instance is Running")
	})

	It("creates Job once instance is woken-set and Running", func() {
		inst := makeRunningInstance("maa-wkready", 0, "10.0.0.2:5555")
		inst.Spec.Suspend = true // spec says suspended, but woken overrides it
		inst.Status.Woken = &redroidv1alpha1.WokenStatus{
			Actor:  "task/task-wkready",
			Reason: "on-demand wake for one-shot task task-wkready",
		}
		task := makeTaskWakeInstance("task-wkready", []string{"maa-wkready"})

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
			WithObjects(inst, task).Build()

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-wkready") // add finalizer
		reconcileTask(r, "task-wkready") // create Job

		jobList := &batchv1.JobList{}
		err := fakeClient.List(context.Background(), jobList)
		Expect(err).NotTo(HaveOccurred(), "list jobs")
		Expect(jobList.Items).To(HaveLen(1), "expected 1 job after instance running")
	})

	It("clears status.woken from instance once Job finishes", func() {
		inst := makeRunningInstance("maa-wkdone", 0, "10.0.0.2:5555")
		inst.Spec.Suspend = true
		inst.Status.Woken = &redroidv1alpha1.WokenStatus{
			Actor:  "task/task-wkdone",
			Reason: "on-demand wake for one-shot task task-wkdone",
		}
		task := makeTaskWakeInstance("task-wkdone", []string{"maa-wkdone"})

		// Create a finished Job manually.
		trueVal := true
		finishedJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "task-wkdone-maa-wkdone",
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

		r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileTask(r, "task-wkdone") // add finalizer
		reconcileTask(r, "task-wkdone") // detect finished job → clear woken

		updatedInst := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "maa-wkdone", Namespace: "default"}, updatedInst)
		Expect(err).NotTo(HaveOccurred(), "get instance")
		Expect(updatedInst.Status.Woken).To(BeNil(), "expected woken to be cleared after Job completion")
	})
})
