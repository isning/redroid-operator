package controller_test

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// helpers

func makeRunningInstance(name string, index int, adb string) *redroidv1alpha1.RedroidInstance {
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
			Phase:      redroidv1alpha1.RedroidInstanceRunning,
			ADBAddress: adb,
			PodName:    "redroid-instance-" + name,
		},
	}
}

func makeTask(name string, instances []string, schedule string, integrations []redroidv1alpha1.IntegrationSpec) *redroidv1alpha1.RedroidTask {
	refs := make([]redroidv1alpha1.InstanceRef, len(instances))
	for i, n := range instances {
		refs[i] = redroidv1alpha1.InstanceRef{Name: n}
	}
	return &redroidv1alpha1.RedroidTask{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: redroidv1alpha1.RedroidTaskSpec{
			Instances:    refs,
			Schedule:     schedule,
			Integrations: integrations,
		},
	}
}

func reconcileTask(t *testing.T, r *controller.RedroidTaskReconciler, name string) ctrl.Result {
	t.Helper()
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	return res
}

func basicIntegration() redroidv1alpha1.IntegrationSpec {
	return redroidv1alpha1.IntegrationSpec{
		Name:    "maa-cli",
		Image:   "maa-cli:latest",
		Command: []string{"maa"},
		Args:    []string{"run", "daily"},
	}
}

// ---- tests ----

// TestRedroidTask_AddsFinalizer verifies finalizer is set on first reconcile.
func TestRedroidTask_AddsFinalizer(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-fin", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-fin")

	updated := &redroidv1alpha1.RedroidTask{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-fin", Namespace: "default"}, updated); err != nil {
		t.Fatalf("get task: %v", err)
	}
	if !controllerutil.ContainsFinalizer(updated, "redroid.isning.moe/task-finalizer") {
		t.Error("expected task finalizer to be set")
	}
}

// TestRedroidTask_CreatesJobPerInstance verifies one Job per instance for one-shot tasks.
func TestRedroidTask_CreatesJobPerInstance(t *testing.T) {
	scheme := newTestScheme(t)
	inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
	task := makeTask("task-job", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst0, inst1, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-job")
	reconcileTask(t, r, "task-job")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobList.Items))
	}
}

// TestRedroidTask_JobHasCorrectPodSpec verifies the redroid sidecar init container is present.
func TestRedroidTask_JobHasCorrectPodSpec(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-spec", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-spec")
	reconcileTask(t, r, "task-spec")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs created")
	}
	podSpec := jobList.Items[0].Spec.Template.Spec
	if len(podSpec.InitContainers) == 0 {
		t.Fatal("expected at least one init container (redroid sidecar)")
	}
	found := false
	for _, ic := range podSpec.InitContainers {
		if ic.Name == "redroid" {
			found = true
			restart := corev1.ContainerRestartPolicyAlways
			if ic.RestartPolicy == nil || *ic.RestartPolicy != restart {
				t.Error("redroid sidecar should have restartPolicy: Always")
			}
		}
	}
	if !found {
		t.Error("expected init container named 'redroid'")
	}
	if len(podSpec.Containers) == 0 {
		t.Fatal("expected at least one main container")
	}
}

// TestRedroidTask_IntegrationEnvVarsInjected verifies ADB_ADDRESS and INSTANCE_INDEX are injected.
func TestRedroidTask_IntegrationEnvVarsInjected(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-5", 5, "127.0.0.1:5555")
	task := makeTask("task-env", []string{"maa-5"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-env")
	reconcileTask(t, r, "task-env")

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
	envMap := map[string]string{}
	for _, e := range containers[0].Env {
		envMap[e.Name] = e.Value
	}
	if envMap["ADB_ADDRESS"] != "127.0.0.1:5555" {
		t.Errorf("expected ADB_ADDRESS=127.0.0.1:5555, got %q", envMap["ADB_ADDRESS"])
	}
	if envMap["INSTANCE_INDEX"] != "5" {
		t.Errorf("expected INSTANCE_INDEX=5, got %q", envMap["INSTANCE_INDEX"])
	}
}

// TestRedroidTask_ConfigMountedAsVolume verifies ConfigFile entries create volumes and mounts.
// It also verifies that multiple Configs referencing different ConfigMaps each get their own
// volume (regression: old code used a per-integration volume name causing collisions).
func TestRedroidTask_ConfigMountedAsVolume(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	integration := redroidv1alpha1.IntegrationSpec{
		Name:    "maa-cli",
		Image:   "maa-cli:latest",
		Command: []string{"maa"},
		Configs: []redroidv1alpha1.ConfigFile{
			{ConfigMapName: "maa-config", Key: "maa-config.json", MountPath: "/etc/maa/maa-config.json"},
			{ConfigMapName: "extra-config", Key: "extra.json", MountPath: "/etc/maa/extra.json"},
		},
	}
	task := makeTask("task-cfg", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{integration})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-cfg")
	reconcileTask(t, r, "task-cfg")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	podSpec := jobList.Items[0].Spec.Template.Spec
	volumeNames := map[string]bool{}
	for _, v := range podSpec.Volumes {
		volumeNames[v.Name] = true
	}
	// Volume names are derived via ConfigMapVolumeName (includes a hash suffix for collision safety).
	for _, cmName := range []string{"maa-config", "extra-config"} {
		wantVol := controller.ConfigMapVolumeName(cmName)
		if !volumeNames[wantVol] {
			t.Errorf("expected volume %q, got volumes: %v", wantVol, volumeNames)
		}
	}
	mountPaths := map[string]bool{}
	if len(podSpec.Containers) > 0 {
		for _, vm := range podSpec.Containers[0].VolumeMounts {
			mountPaths[vm.MountPath] = true
		}
	}
	for _, wantMount := range []string{"/etc/maa/maa-config.json", "/etc/maa/extra.json"} {
		if !mountPaths[wantMount] {
			t.Errorf("expected mount %q, got mounts: %v", wantMount, mountPaths)
		}
	}
}

// TestRedroidTask_CreatesCronJobPerInstance verifies CronJobs created for scheduled tasks.
func TestRedroidTask_CreatesCronJobPerInstance(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-cron", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-cron")
	reconcileTask(t, r, "task-cron")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) != 1 {
		t.Errorf("expected 1 CronJob, got %d", len(cronList.Items))
	}
	if cronList.Items[0].Spec.Schedule != "0 4 * * *" {
		t.Errorf("unexpected schedule: %q", cronList.Items[0].Spec.Schedule)
	}
}

// TestRedroidTask_NoCronJobForOneShotTask verifies no CronJob for tasks without schedule.
func TestRedroidTask_NoCronJobForOneShotTask(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-oneshot", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-oneshot")
	reconcileTask(t, r, "task-oneshot")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) != 0 {
		t.Errorf("expected no CronJobs for one-shot task, got %d", len(cronList.Items))
	}
}

// TestRedroidTask_CronJobSuspendSynced verifies Suspend field is propagated to CronJob.
func TestRedroidTask_CronJobSuspendSynced(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-suspend", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
	task.Spec.Suspend = true

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-suspend")
	reconcileTask(t, r, "task-suspend")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	cj := cronList.Items[0]
	if cj.Spec.Suspend == nil || !*cj.Spec.Suspend {
		t.Error("expected CronJob.Spec.Suspend=true when task.Spec.Suspend=true")
	}
}

// TestRedroidTask_CronJobHistoryLimits verifies success/fail history limits are set.
func TestRedroidTask_CronJobHistoryLimits(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-hist", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-hist")
	reconcileTask(t, r, "task-hist")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	cj := cronList.Items[0]
	if cj.Spec.SuccessfulJobsHistoryLimit == nil || *cj.Spec.SuccessfulJobsHistoryLimit != 3 {
		t.Errorf("expected SuccessfulJobsHistoryLimit=3, got %v", cj.Spec.SuccessfulJobsHistoryLimit)
	}
	if cj.Spec.FailedJobsHistoryLimit == nil || *cj.Spec.FailedJobsHistoryLimit != 3 {
		t.Errorf("expected FailedJobsHistoryLimit=3, got %v", cj.Spec.FailedJobsHistoryLimit)
	}
}

// TestRedroidTask_IgnoreNotFound verifies reconcile does not error on missing resource.
func TestRedroidTask_IgnoreNotFound(t *testing.T) {
	scheme := newTestScheme(t)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidTask{}).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "ghost", Namespace: "default"},
	})
	if err != nil {
		t.Errorf("expected no error for missing task, got: %v", err)
	}
}

// TestIsJobFinished_Complete verifies finished job detection.
func TestIsJobFinished_Complete(t *testing.T) {
	complete := batchv1.Job{
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			},
		},
	}
	if !controller.IsJobFinished(&complete) {
		t.Error("expected IsJobFinished=true for complete job")
	}

	failed := batchv1.Job{
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{Type: batchv1.JobFailed, Status: corev1.ConditionTrue},
			},
		},
	}
	if !controller.IsJobFinished(&failed) {
		t.Error("expected IsJobFinished=true for failed job")
	}

	running := batchv1.Job{}
	if controller.IsJobFinished(&running) {
		t.Error("expected IsJobFinished=false for running job")
	}
}

// TestRedroidTask_OverlayfsVolumesPresent verifies volumes data-base, data-diff, dev-dri in Job pod spec.
func TestRedroidTask_OverlayfsVolumesPresent(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-3", 3, "10.0.0.4:5555")
	task := makeTask("task-ovl", []string{"maa-3"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-ovl")
	reconcileTask(t, r, "task-ovl")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	podSpec := jobList.Items[0].Spec.Template.Spec
	volNames := map[string]bool{}
	for _, v := range podSpec.Volumes {
		volNames[v.Name] = true
	}
	for _, want := range []string{"data-base", "data-diff", "dev-dri"} {
		if !volNames[want] {
			t.Errorf("expected volume %q not found; volumes: %v", want, volNames)
		}
	}
	sidecars := podSpec.InitContainers
	if len(sidecars) == 0 {
		t.Fatal("no init containers")
	}
	var sidecarMounts []string
	for _, ic := range sidecars {
		if ic.Name == "redroid" {
			for _, vm := range ic.VolumeMounts {
				sidecarMounts = append(sidecarMounts, vm.MountPath)
			}
		}
	}
	wantMounts := map[string]bool{"/data-base": true, "/data-diff/3": true, "/dev/dri": true}
	for _, m := range sidecarMounts {
		delete(wantMounts, m)
	}
	if len(wantMounts) > 0 {
		t.Errorf("redroid sidecar missing mounts: %v; got: %v", wantMounts, sidecarMounts)
	}
}

// TestRedroidTask_MultipleIntegrations verifies multiple integration containers.
func TestRedroidTask_MultipleIntegrations(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	integrations := []redroidv1alpha1.IntegrationSpec{
		{Name: "tool-a", Image: "tool-a:latest"},
		{Name: "tool-b", Image: "tool-b:latest"},
		{Name: "tool-c", Image: "tool-c:latest"},
	}
	task := makeTask("task-multi", []string{"maa-0"}, "", integrations)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-multi")
	reconcileTask(t, r, "task-multi")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs")
	}
	containers := jobList.Items[0].Spec.Template.Spec.Containers
	if len(containers) != 3 {
		t.Errorf("expected 3 containers for 3 integrations, got %d", len(containers))
	}
}

// TestRedroidTask_ConcurrencyPolicyForbid verifies CronJob uses ForbidConcurrent policy.
func TestRedroidTask_ConcurrencyPolicyForbid(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-concur", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-concur")
	reconcileTask(t, r, "task-concur")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	if cronList.Items[0].Spec.ConcurrencyPolicy != batchv1.ForbidConcurrent {
		t.Errorf("expected ForbidConcurrent, got %v", cronList.Items[0].Spec.ConcurrencyPolicy)
	}
}

// TestRedroidTask_DeletionRemovesFinalizer verifies that deleting a task strips the
// controller finalizer "redroid.isning.moe/task-finalizer" so the object can be garbage collected.
func TestRedroidTask_DeletionRemovesFinalizer(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-del", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	// Reconcile once: controller adds "redroid.isning.moe/task-finalizer" then creates Job.
	reconcileTask(t, r, "task-del")
	reconcileTask(t, r, "task-del")

	// Delete the task. Fake client sets DeletionTimestamp because finalizer is present.
	current := &redroidv1alpha1.RedroidTask{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-del", Namespace: "default"}, current); err != nil {
		t.Fatalf("get task: %v", err)
	}
	if err := fakeClient.Delete(context.Background(), current); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	reconcileTask(t, r, "task-del")

	// After finalizer removal the fake client deletes the object, or it exists with no finalizer.
	final := &redroidv1alpha1.RedroidTask{}
	err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-del", Namespace: "default"}, final)
	if err == nil && controllerutil.ContainsFinalizer(final, "redroid.isning.moe/task-finalizer") {
		t.Error("expected 'redroid.isning.moe/task-finalizer' to be removed after deletion reconcile")
	}
}

// TestRedroidTask_ResolveInstancesError verifies that referencing a non-existent instance
// causes Reconcile to surface an error.
func TestRedroidTask_ResolveInstancesError(t *testing.T) {
	scheme := newTestScheme(t)
	// Task references "missing-instance" which is never registered.
	task := makeTask("task-bad-ref", []string{"missing-instance"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	// The controller adds the finalizer and then immediately tries to resolve instances in the
	// same Reconcile call (no early return after AddFinalizer). So the first call returns an error.
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "task-bad-ref", Namespace: "default"},
	})
	if err == nil {
		t.Error("expected error when instance reference cannot be resolved")
	}
}

// TestRedroidTask_CronJobUpdatedOnSpecChange verifies that when the task spec changes
// (schedule, timezone, suspend) the controller patches the existing CronJob accordingly.
func TestRedroidTask_CronJobUpdatedOnSpecChange(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
	task := makeTask("task-upd", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-upd")
	reconcileTask(t, r, "task-upd") // CronJob now exists

	// Mutate the task schedule + set timezone.
	current := &redroidv1alpha1.RedroidTask{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-upd", Namespace: "default"}, current); err != nil {
		t.Fatalf("get task: %v", err)
	}
	patch := current.DeepCopy()
	patch.Spec.Schedule = "0 5 * * *"
	patch.Spec.Timezone = "UTC"
	patch.Spec.Suspend = true
	if err := fakeClient.Update(context.Background(), patch); err != nil {
		t.Fatalf("update task: %v", err)
	}

	reconcileTask(t, r, "task-upd")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("expected CronJob")
	}
	cj := cronList.Items[0]
	if cj.Spec.Schedule != "0 5 * * *" {
		t.Errorf("expected updated schedule '0 5 * * *', got %q", cj.Spec.Schedule)
	}
	if cj.Spec.TimeZone == nil || *cj.Spec.TimeZone != "UTC" {
		t.Errorf("expected timezone 'UTC' in updated CronJob")
	}
	if cj.Spec.Suspend == nil || !*cj.Spec.Suspend {
		t.Errorf("expected CronJob to be suspended after task.Spec.Suspend=true")
	}
}

// assertPerInstanceVolumesMerged verifies that per-instance volumes and mounts
// are correctly merged into a PodSpec: the volume 'inst-secret' must be present
// with SecretName "instance-secret" (instance definition wins over task-level),
// and the mount '/etc/secret' must appear on every integration container.
func assertPerInstanceVolumesMerged(t *testing.T, podSpec corev1.PodSpec) {
	t.Helper()

	// Volume presence, uniqueness, and instance-wins-over-task precedence.
	var foundVol *corev1.Volume
	count := 0
	for i := range podSpec.Volumes {
		if podSpec.Volumes[i].Name == "inst-secret" {
			count++
			foundVol = &podSpec.Volumes[i]
		}
	}
	if count == 0 {
		t.Fatalf("expected per-instance volume 'inst-secret', got volumes: %v", podSpec.Volumes)
	}
	if count > 1 {
		t.Errorf("expected exactly one volume named 'inst-secret', got %d", count)
	}
	if foundVol.Secret == nil || foundVol.Secret.SecretName != "instance-secret" {
		t.Errorf("instance volume should win: expected SecretName 'instance-secret', got %+v", foundVol.VolumeSource)
	}

	// Mount present on every integration container.
	if len(podSpec.Containers) == 0 {
		t.Fatal("no integration containers")
	}
	for _, c := range podSpec.Containers {
		found := false
		for _, vm := range c.VolumeMounts {
			if vm.MountPath == "/etc/secret" {
				if vm.Name != "inst-secret" {
					t.Errorf("container %q: mount '/etc/secret' has wrong Name: got %q, want 'inst-secret'", c.Name, vm.Name)
				}
				if !vm.ReadOnly {
					t.Errorf("container %q: mount '/etc/secret' should be ReadOnly", c.Name)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("container %q missing per-instance mount '/etc/secret'; mounts: %v", c.Name, c.VolumeMounts)
		}
	}
}

// TestRedroidTask_PerInstanceVolumes verifies that per-instance volumes and
// volumeMounts from InstanceRef are merged into the generated Job's PodSpec.
func TestRedroidTask_PerInstanceVolumes(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")

	task := &redroidv1alpha1.RedroidTask{
		ObjectMeta: metav1.ObjectMeta{Name: "task-inst-vol", Namespace: "default"},
		Spec: redroidv1alpha1.RedroidTaskSpec{
			Instances: []redroidv1alpha1.InstanceRef{
				{
					Name: "maa-0",
					Volumes: []corev1.Volume{
						{
							Name: "inst-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{SecretName: "instance-secret"},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "inst-secret", MountPath: "/etc/secret", ReadOnly: true},
					},
				},
			},
			// Add a conflicting task-level volume with the same name — instance must win.
			Volumes: []corev1.Volume{
				{
					Name: "inst-secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{SecretName: "task-secret"},
					},
				},
			},
			Integrations: []redroidv1alpha1.IntegrationSpec{basicIntegration(), {
				Name:  "second-integration",
				Image: "sidecar:latest",
			}},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-inst-vol")
	reconcileTask(t, r, "task-inst-vol")

	jobList := &batchv1.JobList{}
	if err := fakeClient.List(context.Background(), jobList); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobList.Items) == 0 {
		t.Fatal("no jobs created")
	}
	podSpec := jobList.Items[0].Spec.Template.Spec
	assertPerInstanceVolumesMerged(t, podSpec)
}

// TestRedroidTask_PerInstanceVolumes_CronJob mirrors TestRedroidTask_PerInstanceVolumes
// but targets the CronJob reconciliation path (non-empty Schedule).
func TestRedroidTask_PerInstanceVolumes_CronJob(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")

	task := &redroidv1alpha1.RedroidTask{
		ObjectMeta: metav1.ObjectMeta{Name: "task-inst-vol-cron", Namespace: "default"},
		Spec: redroidv1alpha1.RedroidTaskSpec{
			Schedule: "0 4 * * *", // non-empty → controller creates a CronJob
			Instances: []redroidv1alpha1.InstanceRef{
				{
					Name: "maa-0",
					Volumes: []corev1.Volume{
						{
							Name: "inst-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{SecretName: "instance-secret"},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "inst-secret", MountPath: "/etc/secret", ReadOnly: true},
					},
				},
			},
			// Conflicting task-level volume — instance must win.
			Volumes: []corev1.Volume{
				{
					Name: "inst-secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{SecretName: "task-secret"},
					},
				},
			},
			Integrations: []redroidv1alpha1.IntegrationSpec{basicIntegration(), {
				Name:  "second-integration",
				Image: "sidecar:latest",
			}},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
		WithObjects(inst, task).Build()

	r := &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme}
	reconcileTask(t, r, "task-inst-vol-cron")
	reconcileTask(t, r, "task-inst-vol-cron")

	cronList := &batchv1.CronJobList{}
	if err := fakeClient.List(context.Background(), cronList); err != nil {
		t.Fatalf("list cronjobs: %v", err)
	}
	if len(cronList.Items) == 0 {
		t.Fatal("no cronjobs created")
	}
	podSpec := cronList.Items[0].Spec.JobTemplate.Spec.Template.Spec
	assertPerInstanceVolumesMerged(t, podSpec)
}
