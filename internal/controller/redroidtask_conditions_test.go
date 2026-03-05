package controller_test

import (
	"context"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// ── Active condition ─────────────────────────────────────────────────────────

// TestTaskCondition_Active_NoJobs verifies Active=False with Reason=NoActiveJobs
// when all jobs are finished.
func TestTaskCondition_Active_NoJobs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("cond-active-none", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name) // adds finalizer
	reconcileTask(t, r, task.Name) // creates Job

	// Mark the job as Complete so it is no longer tracked as active.
	jobList := &batchv1.JobList{}
	if err := fc.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("expected a job to be created")
	}
	job := jobList.Items[0].DeepCopy()
	job.Status.Conditions = []batchv1.JobCondition{
		{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
	}
	if err := fc.Status().Update(context.Background(), job); err != nil {
		t.Fatalf("update job status: %v", err)
	}

	reconcileTask(t, r, task.Name) // patchTaskStatus with no active jobs

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Active")
	if cond == nil {
		t.Fatal("Active condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Active=False, got %v", cond.Status)
	}
	if cond.Reason != "NoActiveJobs" {
		t.Errorf("expected Reason=NoActiveJobs, got %q", cond.Reason)
	}
}

// TestTaskCondition_Active_WithRunningJobs verifies Active=True with Reason=JobsActive
// and job names listed in the message while a job is still running.
func TestTaskCondition_Active_WithRunningJobs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("cond-active-run", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name) // job created, not yet finished

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Active")
	if cond == nil {
		t.Fatal("Active condition not found")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Active=True while job running, got %v", cond.Status)
	}
	if cond.Reason != "JobsActive" {
		t.Errorf("expected Reason=JobsActive, got %q", cond.Reason)
	}
	// Job name should be embedded in the message.
	expectedJobName := task.Name + "-maa-0"
	if !strings.Contains(cond.Message, expectedJobName) {
		t.Errorf("expected job name %q in Active message, got %q", expectedJobName, cond.Message)
	}
}

// ── Failed condition ─────────────────────────────────────────────────────────

// TestTaskCondition_Failed_WithFailedJob verifies Failed=True when a job reports
// a Failed condition, and that the job name plus reason appear in the message.
func TestTaskCondition_Failed_WithFailedJob(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("cond-fail-job", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name) // job created

	// Simulate job failure with a reason.
	jobList := &batchv1.JobList{}
	if err := fc.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("expected a job to be created")
	}
	job := jobList.Items[0].DeepCopy()
	job.Status.Conditions = []batchv1.JobCondition{
		{
			Type:    batchv1.JobFailed,
			Status:  corev1.ConditionTrue,
			Reason:  "BackoffLimitExceeded",
			Message: "Job has reached the specified backoff limit",
		},
	}
	if err := fc.Status().Update(context.Background(), job); err != nil {
		t.Fatalf("update job status: %v", err)
	}

	reconcileTask(t, r, task.Name)

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Failed")
	if cond == nil {
		t.Fatal("Failed condition not found")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Failed=True, got %v", cond.Status)
	}
	if cond.Reason != "JobsFailed" {
		t.Errorf("expected Reason=JobsFailed, got %q", cond.Reason)
	}
	if !strings.Contains(cond.Message, job.Name) {
		t.Errorf("expected job name %q in Failed message, got %q", job.Name, cond.Message)
	}
	if !strings.Contains(cond.Message, "BackoffLimitExceeded") {
		t.Errorf("expected reason 'BackoffLimitExceeded' in Failed message, got %q", cond.Message)
	}
	if !strings.Contains(cond.Message, "backoff limit") {
		t.Errorf("expected job failure message in condition message, got %q", cond.Message)
	}
}

// TestTaskCondition_Failed_NoFailedJobs verifies Failed=False when all jobs succeed.
func TestTaskCondition_Failed_NoFailedJobs(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("cond-fail-none", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name)

	// Mark the job as Complete (success).
	jobList := &batchv1.JobList{}
	if err := fc.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	job := jobList.Items[0].DeepCopy()
	job.Status.Conditions = []batchv1.JobCondition{
		{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
	}
	if err := fc.Status().Update(context.Background(), job); err != nil {
		t.Fatalf("update job status: %v", err)
	}

	reconcileTask(t, r, task.Name)

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Failed")
	if cond == nil {
		t.Fatal("Failed condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Failed=False when job succeeded, got %v", cond.Status)
	}
	if cond.Reason != "NoFailedJobs" {
		t.Errorf("expected Reason=NoFailedJobs, got %q", cond.Reason)
	}
}

// ── Complete condition ───────────────────────────────────────────────────────

// TestTaskCondition_Complete_AllSucceeded verifies Complete=True with Reason=AllJobsSucceeded
// and a count in the message when all jobs for a one-shot task have completed.
func TestTaskCondition_Complete_AllSucceeded(t *testing.T) {
	scheme := newTestScheme(t)
	inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
	task := makeTask("cond-complete-all", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst0, inst1, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name) // 2 jobs created

	// Mark both jobs as Complete.
	jobList := &batchv1.JobList{}
	if err := fc.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobList.Items))
	}
	for i := range jobList.Items {
		j := jobList.Items[i].DeepCopy()
		j.Status.Conditions = []batchv1.JobCondition{
			{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
		}
		if err := fc.Status().Update(context.Background(), j); err != nil {
			t.Fatalf("update job[%d] status: %v", i, err)
		}
	}

	reconcileTask(t, r, task.Name)

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Complete")
	if cond == nil {
		t.Fatal("Complete condition not found")
	}
	if cond.Status != metav1.ConditionTrue {
		t.Errorf("expected Complete=True, got %v", cond.Status)
	}
	if cond.Reason != "AllJobsSucceeded" {
		t.Errorf("expected Reason=AllJobsSucceeded, got %q", cond.Reason)
	}
	// Message should mention the count ("2").
	if !strings.Contains(cond.Message, "2") {
		t.Errorf("expected succeeded count in Complete message, got %q", cond.Message)
	}
}

// TestTaskCondition_Complete_Partial verifies Complete=False with a "N/M" progress
// message when only some jobs have finished for a multi-instance task.
func TestTaskCondition_Complete_Partial(t *testing.T) {
	scheme := newTestScheme(t)
	inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
	task := makeTask("cond-complete-partial", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}, &batchv1.Job{}).
		WithObjects(inst0, inst1, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name) // 2 jobs created

	// Mark only first job as complete; second stays running.
	jobList := &batchv1.JobList{}
	if err := fc.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobList.Items))
	}
	first := jobList.Items[0].DeepCopy()
	first.Status.Conditions = []batchv1.JobCondition{
		{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
	}
	if err := fc.Status().Update(context.Background(), first); err != nil {
		t.Fatalf("update first job status: %v", err)
	}
	// second job left with no conditions (still active)

	reconcileTask(t, r, task.Name)

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}
	cond := findTaskCondition(result.Status.Conditions, "Complete")
	if cond == nil {
		t.Fatal("Complete condition not found")
	}
	if cond.Status != metav1.ConditionFalse {
		t.Errorf("expected Complete=False while one job is pending, got %v", cond.Status)
	}
	// Message should express fractional progress.
	if !strings.Contains(cond.Message, "1") || !strings.Contains(cond.Message, "2") {
		t.Errorf("expected '1/2' style progress in message, got %q", cond.Message)
	}
}

// ── CronJob (scheduled) task conditions ─────────────────────────────────────

// TestTaskCondition_Scheduled_CronJob verifies that a CronJob-based task gets
// stable "Scheduled" reason on both Complete and Failed conditions, so users
// understand they should look at the owned Job history instead.
func TestTaskCondition_Scheduled_CronJob(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("cond-cron-sched", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fc := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()
	r := &controller.RedroidTaskReconciler{Client: fc, Scheme: scheme}

	reconcileTask(t, r, task.Name)
	reconcileTask(t, r, task.Name)

	result := &redroidv1alpha1.RedroidTask{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: task.Name, Namespace: "default"}, result); err != nil {
		t.Fatalf("get task: %v", err)
	}

	for _, condType := range []string{"Complete", "Failed"} {
		cond := findTaskCondition(result.Status.Conditions, condType)
		if cond == nil {
			t.Fatalf("%s condition not found", condType)
		}
		if cond.Reason != "Scheduled" {
			t.Errorf("%s: expected Reason=Scheduled for CronJob task, got %q", condType, cond.Reason)
		}
		if !strings.Contains(strings.ToLower(cond.Message), "cronjob") {
			t.Errorf("%s: expected 'cronjob' in message, got %q", condType, cond.Message)
		}
	}
}
