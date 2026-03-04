package controller_test

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// TestRedroidTask_CustomBackoffLimit verifies a custom BackoffLimit is applied to the Job.
func TestRedroidTask_CustomBackoffLimit(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-bl", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	bl := int32(5)
	task.Spec.BackoffLimit = &bl

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-bl")
	reconcileTask(t, r, "task-bl")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	if jobList.Items[0].Spec.BackoffLimit == nil || *jobList.Items[0].Spec.BackoffLimit != 5 {
		t.Errorf("expected BackoffLimit=5, got %v", jobList.Items[0].Spec.BackoffLimit)
	}
}

// TestRedroidTask_TTLSecondsAfterFinished verifies TTL is applied to one-shot Jobs.
func TestRedroidTask_TTLSecondsAfterFinished(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-ttl", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	ttl := int32(300)
	task.Spec.TTLSecondsAfterFinished = &ttl

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-ttl")
	reconcileTask(t, r, "task-ttl")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	if jobList.Items[0].Spec.TTLSecondsAfterFinished == nil || *jobList.Items[0].Spec.TTLSecondsAfterFinished != 300 {
		t.Errorf("expected TTLSecondsAfterFinished=300, got %v", jobList.Items[0].Spec.TTLSecondsAfterFinished)
	}
}

// TestRedroidTask_Timezone verifies timezone is applied to CronJob.
func TestRedroidTask_Timezone(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-tz", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	task.Spec.Timezone = "Asia/Shanghai"

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-tz")
	reconcileTask(t, r, "task-tz")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	if cronList.Items[0].Spec.TimeZone == nil || *cronList.Items[0].Spec.TimeZone != "Asia/Shanghai" {
		t.Errorf("expected Timezone=Asia/Shanghai, got %v", cronList.Items[0].Spec.TimeZone)
	}
}

// TestRedroidTask_ObservedGeneration verifies ObservedGeneration is set in task status.
func TestRedroidTask_ObservedGeneration(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-gen", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-gen")
	reconcileTask(t, r, "task-gen")

	result := &redroidv1alpha1.RedroidTask{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-gen", Namespace: "default"}, result); err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Status.ObservedGeneration != result.Generation {
		t.Errorf("expected ObservedGeneration=%d, got %d", result.Generation, result.Status.ObservedGeneration)
	}
}

// TestRedroidTask_IntegrationWorkingDir verifies WorkingDir is applied to integration containers.
func TestRedroidTask_IntegrationWorkingDir(t *testing.T) {
	scheme := newTestScheme(t)
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

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-wd")
	reconcileTask(t, r, "task-wd")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	containers := jobList.Items[0].Spec.Template.Spec.Containers
	if len(containers) == 0 {
		t.Fatal("no containers")
	}
	if containers[0].WorkingDir != "/workspace" {
		t.Errorf("expected WorkingDir=/workspace, got %q", containers[0].WorkingDir)
	}
}

// TestRedroidTask_StartingDeadlineSeconds verifies StartingDeadlineSeconds is applied to CronJob.
func TestRedroidTask_StartingDeadlineSeconds(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-sds", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	sds := int64(180)
	task.Spec.StartingDeadlineSeconds = &sds

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-sds")
	reconcileTask(t, r, "task-sds")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	if cronList.Items[0].Spec.StartingDeadlineSeconds == nil || *cronList.Items[0].Spec.StartingDeadlineSeconds != 180 {
		t.Errorf("expected StartingDeadlineSeconds=180, got %v", cronList.Items[0].Spec.StartingDeadlineSeconds)
	}
}

// TestRedroidTask_TaskImagePullSecrets verifies task ImagePullSecrets appear in the Pod spec.
func TestRedroidTask_TaskImagePullSecrets(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-ips", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	task.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "my-registry-secret"}}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-ips")
	reconcileTask(t, r, "task-ips")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	pullSecrets := jobList.Items[0].Spec.Template.Spec.ImagePullSecrets
	found := false
	for _, s := range pullSecrets {
		if s.Name == "my-registry-secret" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ImagePullSecret 'my-registry-secret' in pod spec, got: %v", pullSecrets)
	}
}

// TestRedroidTask_ServiceAccountName verifies that ServiceAccountName set on an
// IntegrationSpec is propagated to the Job's PodSpec.
func TestRedroidTask_ServiceAccountName(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	integ := basicIntegration()
	integ.ServiceAccountName = "my-task-sa"
	task := makeTask("task-sa", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{integ})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-sa")
	reconcileTask(t, r, "task-sa")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs created")
	}
	got := jobList.Items[0].Spec.Template.Spec.ServiceAccountName
	if got != "my-task-sa" {
		t.Errorf("ServiceAccountName: want %q, got %q", "my-task-sa", got)
	}
}
