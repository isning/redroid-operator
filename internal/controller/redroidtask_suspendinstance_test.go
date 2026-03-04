package controller_test

// Tests for task.Spec.SuspendInstance — the mechanism where a one-shot Task
// temporarily stops the referenced RedroidInstance before running and restores
// it afterwards to prevent overlayfs conflicts.

import (
	"context"
	"testing"

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

// TestRedroidTask_SuspendInstanceSetsTempSuspend verifies that when a Job does
// not yet exist the task controller sets status.suspended on the
// instance before creating the Job, then requeues.
func TestRedroidTask_SuspendInstanceSetsTempSuspend(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-sus", 0, "10.0.0.1:5555")
	task := makeTaskSuspendInstance("task-sus", []string{"maa-sus"})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}

	// 1st reconcile: adds finalizer.
	reconcileTask(t, r, "task-sus")
	// 2nd reconcile: should set suspended on instance and requeue.
	res := reconcileTask(t, r, "task-sus")

	if res.RequeueAfter == 0 {
		t.Error("expected RequeueAfter > 0 while waiting for instance to stop")
	}

	// Instance should now have suspended set.
	updatedInst := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "maa-sus", Namespace: "default"}, updatedInst); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if updatedInst.Status.Suspended == nil {
		t.Fatal("expected suspended to be set on instance")
	}
	if updatedInst.Status.Suspended.Actor != "task/task-sus" {
		t.Errorf("expected actor 'task/task-sus', got %q", updatedInst.Status.Suspended.Actor)
	}

	// Job must NOT have been created yet.
	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 0 {
		t.Errorf("expected 0 jobs (instance not stopped yet), got %d", len(jobList.Items))
	}
}

// TestRedroidTask_SuspendInstanceWaitsForStop verifies that the controller
// keeps requeueing without creating a Job when suspended is already set
// but the instance has not yet reached phase=Stopped.
func TestRedroidTask_SuspendInstanceWaitsForStop(t *testing.T) {
	scheme := newTestScheme(t)
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
	reconcileTask(t, r, "task-waits") // add finalizer

	res := reconcileTask(t, r, "task-waits") // should requeue, no job
	if res.RequeueAfter == 0 {
		t.Error("expected RequeueAfter > 0 while instance is still Running")
	}

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 0 {
		t.Errorf("job must not be created before instance is Stopped; got %d job(s)", len(jobList.Items))
	}
}

// TestRedroidTask_SuspendInstanceCreatesJobWhenStopped verifies that the Job is
// created once the instance is both suspended-set AND phase=Stopped.
func TestRedroidTask_SuspendInstanceCreatesJobWhenStopped(t *testing.T) {
	scheme := newTestScheme(t)
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
	reconcileTask(t, r, "task-ready") // add finalizer
	reconcileTask(t, r, "task-ready") // create Job

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 1 {
		t.Errorf("expected 1 job after instance stopped, got %d", len(jobList.Items))
	}
}

// TestRedroidTask_SuspendInstanceClearsAfterJobDone verifies that
// status.suspended is cleared from the instance once the Job finishes.
func TestRedroidTask_SuspendInstanceClearsAfterJobDone(t *testing.T) {
	scheme := newTestScheme(t)
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
			APIVersion: "redroid.io/v1alpha1",
			Kind:       "RedroidTask",
			Name:       task.Name,
			Controller: &trueVal,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task, finishedJob).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-done") // add finalizer
	reconcileTask(t, r, "task-done") // detect finished job → clear suspended

	updatedInst := &redroidv1alpha1.RedroidInstance{}
	if err := fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "maa-done", Namespace: "default"}, updatedInst); err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if updatedInst.Status.Suspended != nil {
		t.Errorf("expected suspended to be cleared after Job completion, got %+v",
			updatedInst.Status.Suspended)
	}
}
