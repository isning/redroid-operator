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
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func reconcileTask(r *controller.RedroidTaskReconciler, name string) ctrl.Result {
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
	})
	// In Ginkgo, we can use Expect here for safety within helper if called synchronously
	Expect(err).ToNot(HaveOccurred(), "Reconcile returned error")
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

// assertPerInstanceVolumesMerged verifies that per-instance volumes and mounts
// are correctly merged into a PodSpec: the volume 'inst-secret' must be present
// with SecretName "instance-secret" (instance definition wins over task-level),
// and the mount '/etc/secret' must appear on every integration container.
func assertPerInstanceVolumesMerged(podSpec corev1.PodSpec) {
	// Volume presence, uniqueness, and instance-wins-over-task precedence.
	var foundVol *corev1.Volume
	count := 0
	for i := range podSpec.Volumes {
		if podSpec.Volumes[i].Name == "inst-secret" {
			count++
			foundVol = &podSpec.Volumes[i]
		}
	}
	Expect(count).To(Equal(1), "expected exactly one volume named 'inst-secret', got volumes: %v", podSpec.Volumes)
	Expect(foundVol.Secret).NotTo(BeNil())
	Expect(foundVol.Secret.SecretName).To(Equal("instance-secret"), "instance volume should win: expected SecretName 'instance-secret', got %+v", foundVol.VolumeSource)

	// Mount present on every integration container.
	Expect(podSpec.Containers).ToNot(BeEmpty(), "no integration containers")
	for _, c := range podSpec.Containers {
		found := false
		for _, vm := range c.VolumeMounts {
			if vm.MountPath == "/etc/secret" {
				Expect(vm.Name).To(Equal("inst-secret"), "container %q: mount '/etc/secret' has wrong Name: got %q, want 'inst-secret'", c.Name, vm.Name)
				Expect(vm.ReadOnly).To(BeTrue(), "container %q: mount '/etc/secret' should be ReadOnly", c.Name)
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "container %q missing per-instance mount '/etc/secret'; mounts: %v", c.Name, c.VolumeMounts)
	}
}

// ---- tests ----

var _ = Describe("RedroidTask Controller", func() {
	var (
		fakeClient client.Client
		r          *controller.RedroidTaskReconciler
		scheme     = newTestScheme()
	)

	BeforeEach(func() {
		// Initialize with empty client builder. Each test context will build its own client if needed,
		// or we can structure the objects differently. Given the previous structure, let's keep it close:
	})

	Context("Reconciling a RedroidTask", func() {
		It("Adds a finalizer to the task", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-fin", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-fin")

			updated := &redroidv1alpha1.RedroidTask{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-fin", Namespace: "default"}, updated)
			Expect(err).NotTo(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(updated, "redroid.isning.moe/task-finalizer")).To(BeTrue(), "expected task finalizer to be set")
		})

		It("Creates one Job per instance for one-shot tasks", func() {
			inst0 := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			inst1 := makeRunningInstance("maa-1", 1, "10.0.0.2:5555")
			task := makeTask("task-job", []string{"maa-0", "maa-1"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst0, inst1, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-job")
			reconcileTask(r, "task-job")

			fakeRec := r.Recorder.(*record.FakeRecorder)
			// Wait for two jobs to be created
			Expect(fakeRec.Events).To(Receive(ContainSubstring("CreatedJob")))
			Expect(fakeRec.Events).To(Receive(ContainSubstring("CreatedJob")))

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).To(HaveLen(2), "expected 2 jobs")
		})

		It("Injects the redroid sidecar init container with Correct PodSpec", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-spec", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-spec")
			reconcileTask(r, "task-spec")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty(), "no jobs created")

			podSpec := jobList.Items[0].Spec.Template.Spec
			Expect(podSpec.InitContainers).NotTo(BeEmpty(), "expected at least one init container (redroid sidecar)")

			found := false
			for _, ic := range podSpec.InitContainers {
				if ic.Name == "redroid" {
					found = true
					restart := corev1.ContainerRestartPolicyAlways
					Expect(ic.RestartPolicy).NotTo(BeNil())
					Expect(*ic.RestartPolicy).To(Equal(restart), "redroid sidecar should have restartPolicy: Always")
					Expect(ic.ReadinessProbe).NotTo(BeNil())
					Expect(ic.ReadinessProbe.Exec.Command).To(ContainElement(ContainSubstring("sys.boot_completed")))
				}
			}
			Expect(found).To(BeTrue(), "expected init container named 'redroid'")
			Expect(podSpec.Containers).NotTo(BeEmpty(), "expected at least one main container")
		})

		It("Injects ADB_ADDRESS and INSTANCE_INDEX env vars into integration", func() {
			inst := makeRunningInstance("maa-5", 5, "127.0.0.1:5555")
			task := makeTask("task-env", []string{"maa-5"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-env")
			reconcileTask(r, "task-env")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

			containers := jobList.Items[0].Spec.Template.Spec.Containers
			Expect(containers).NotTo(BeEmpty(), "no containers")

			envMap := map[string]string{}
			for _, e := range containers[0].Env {
				envMap[e.Name] = e.Value
			}
			Expect(envMap["ADB_ADDRESS"]).To(Equal("127.0.0.1:5555"))
			Expect(envMap["INSTANCE_INDEX"]).To(Equal("5"))
		})

		It("Mounts ConfigFiles as Volumes properly", func() {
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

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-cfg")
			reconcileTask(r, "task-cfg")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty(), "no jobs")

			podSpec := jobList.Items[0].Spec.Template.Spec
			volumeNames := map[string]bool{}
			for _, v := range podSpec.Volumes {
				volumeNames[v.Name] = true
			}
			// Volume names are derived via ConfigMapVolumeName (includes a hash suffix for collision safety).
			for _, cmName := range []string{"maa-config", "extra-config"} {
				wantVol := controller.ConfigMapVolumeName(cmName)
				Expect(volumeNames[wantVol]).To(BeTrue(), "expected volume %q, got volumes: %v", wantVol, volumeNames)
			}

			mountPaths := map[string]bool{}
			if len(podSpec.Containers) > 0 {
				for _, vm := range podSpec.Containers[0].VolumeMounts {
					mountPaths[vm.MountPath] = true
				}
			}
			for _, wantMount := range []string{"/etc/maa/maa-config.json", "/etc/maa/extra.json"} {
				Expect(mountPaths[wantMount]).To(BeTrue(), "expected mount %q, got mounts: %v", wantMount, mountPaths)
			}
		})

		It("Creates CronJob per instance for scheduled tasks", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-cron", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-cron")
			reconcileTask(r, "task-cron")

			fakeRec := r.Recorder.(*record.FakeRecorder)
			Expect(fakeRec.Events).To(Receive(ContainSubstring("CreatedCronJob")))

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).To(HaveLen(1))
			Expect(cronList.Items[0].Spec.Schedule).To(Equal("0 4 * * *"))
		})

		It("Does not create CronJob for one-shot tasks", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-oneshot", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-oneshot")
			reconcileTask(r, "task-oneshot")

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).To(BeEmpty())
		})

		It("Syncs CronJob Suspend field", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-suspend", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})
			task.Spec.Suspend = true

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-suspend")
			reconcileTask(r, "task-suspend")

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).NotTo(BeEmpty())
			cj := cronList.Items[0]
			Expect(cj.Spec.Suspend).NotTo(BeNil())
			Expect(*cj.Spec.Suspend).To(BeTrue())
		})

		It("Sets CronJob history limits to 3", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-hist", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-hist")
			reconcileTask(r, "task-hist")

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).NotTo(BeEmpty())
			cj := cronList.Items[0]
			Expect(cj.Spec.SuccessfulJobsHistoryLimit).NotTo(BeNil())
			Expect(*cj.Spec.SuccessfulJobsHistoryLimit).To(Equal(int32(3)))
			Expect(cj.Spec.FailedJobsHistoryLimit).NotTo(BeNil())
			Expect(*cj.Spec.FailedJobsHistoryLimit).To(Equal(int32(3)))
		})

		It("Ignores NotFound gracefully", func() {
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidTask{}).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			_, err := r.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "ghost", Namespace: "default"},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Includes OverlayfsVolumes in Job pod spec properly", func() {
			inst := makeRunningInstance("maa-3", 3, "10.0.0.4:5555")
			task := makeTask("task-ovl", []string{"maa-3"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-ovl")
			reconcileTask(r, "task-ovl")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty())

			podSpec := jobList.Items[0].Spec.Template.Spec
			volNames := map[string]bool{}
			for _, v := range podSpec.Volumes {
				volNames[v.Name] = true
			}
			for _, want := range []string{"data-base", "data-diff", "dev-dri"} {
				Expect(volNames[want]).To(BeTrue(), "expected volume %q not found; volumes: %v", want, volNames)
			}

			sidecars := podSpec.InitContainers
			Expect(sidecars).NotTo(BeEmpty())

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
			Expect(wantMounts).To(BeEmpty(), "redroid sidecar missing mounts: %v; got: %v", wantMounts, sidecarMounts)
		})

		It("Handles multiple Integrations", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			integrations := []redroidv1alpha1.IntegrationSpec{
				{Name: "tool-a", Image: "tool-a:latest"},
				{Name: "tool-b", Image: "tool-b:latest"},
				{Name: "tool-c", Image: "tool-c:latest"},
			}
			task := makeTask("task-multi", []string{"maa-0"}, "", integrations)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-multi")
			reconcileTask(r, "task-multi")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty())

			containers := jobList.Items[0].Spec.Template.Spec.Containers
			Expect(containers).To(HaveLen(3))
		})

		It("Uses ForbidConcurrent policy for CronJob", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-concur", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-concur")
			reconcileTask(r, "task-concur")

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).NotTo(BeEmpty())
			Expect(cronList.Items[0].Spec.ConcurrencyPolicy).To(Equal(batchv1.ForbidConcurrent))
		})

		It("Removes finalizer on task deletion", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-del", []string{"maa-0"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			// Reconcile once: controller adds "redroid.isning.moe/task-finalizer" then creates Job.
			reconcileTask(r, "task-del")
			reconcileTask(r, "task-del")

			// Delete the task. Fake client sets DeletionTimestamp because finalizer is present.
			current := &redroidv1alpha1.RedroidTask{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-del", Namespace: "default"}, current)
			Expect(err).NotTo(HaveOccurred())

			err = fakeClient.Delete(context.Background(), current)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, "task-del")

			// After finalizer removal the fake client deletes the object, or it exists with no finalizer.
			final := &redroidv1alpha1.RedroidTask{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-del", Namespace: "default"}, final)
			if err == nil {
				Expect(controllerutil.ContainsFinalizer(final, "redroid.isning.moe/task-finalizer")).To(BeFalse(), "expected 'redroid.isning.moe/task-finalizer' to be removed after deletion reconcile")
			}
		})

		It("Returns error if referencing non-existent instance", func() {
			// Task references "missing-instance" which is never registered.
			task := makeTask("task-bad-ref", []string{"missing-instance"}, "", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			// The controller adds the finalizer and then immediately tries to resolve instances in the
			// same Reconcile call. So the first call returns an error.
			_, err := r.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "task-bad-ref", Namespace: "default"},
			})
			Expect(err).To(HaveOccurred(), "expected error when instance reference cannot be resolved")
		})

		It("Updates CronJob on Spec change", func() {
			inst := makeRunningInstance("maa-0", 0, "10.0.0.1:5555")
			task := makeTask("task-upd", []string{"maa-0"}, "0 4 * * *", []redroidv1alpha1.IntegrationSpec{basicIntegration()})

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-upd")
			reconcileTask(r, "task-upd") // CronJob now exists

			// Mutate the task schedule + set timezone.
			current := &redroidv1alpha1.RedroidTask{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "task-upd", Namespace: "default"}, current)
			Expect(err).NotTo(HaveOccurred())

			patch := current.DeepCopy()
			patch.Spec.Schedule = "0 5 * * *"
			patch.Spec.Timezone = "UTC"
			patch.Spec.Suspend = true
			err = fakeClient.Update(context.Background(), patch)
			Expect(err).NotTo(HaveOccurred())

			reconcileTask(r, "task-upd")

			cronList := &batchv1.CronJobList{}
			err = fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).NotTo(BeEmpty())

			cj := cronList.Items[0]
			Expect(cj.Spec.Schedule).To(Equal("0 5 * * *"), "expected updated schedule")
			Expect(cj.Spec.TimeZone).NotTo(BeNil())
			Expect(*cj.Spec.TimeZone).To(Equal("UTC"), "expected timezone 'UTC' in updated CronJob")
			Expect(cj.Spec.Suspend).NotTo(BeNil())
			Expect(*cj.Spec.Suspend).To(BeTrue(), "expected CronJob to be suspended after task.Spec.Suspend=true")
		})

		It("Merges PerInstanceVolumes for one-shot Jobs", func() {
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

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-inst-vol")
			reconcileTask(r, "task-inst-vol")

			jobList := &batchv1.JobList{}
			err := fakeClient.List(context.Background(), jobList)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobList.Items).NotTo(BeEmpty())

			podSpec := jobList.Items[0].Spec.Template.Spec
			assertPerInstanceVolumesMerged(podSpec)
		})

		It("Merges PerInstanceVolumes for CronJobs", func() {
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

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}, &redroidv1alpha1.RedroidTask{}).
				WithObjects(inst, task).Build()

			r = &controller.RedroidTaskReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileTask(r, "task-inst-vol-cron")
			reconcileTask(r, "task-inst-vol-cron")

			cronList := &batchv1.CronJobList{}
			err := fakeClient.List(context.Background(), cronList)
			Expect(err).NotTo(HaveOccurred())
			Expect(cronList.Items).NotTo(BeEmpty())

			podSpec := cronList.Items[0].Spec.JobTemplate.Spec.Template.Spec
			assertPerInstanceVolumesMerged(podSpec)
		})
	})
})

var _ = Describe("IsJobFinished", func() {
	It("Returns true for a complete job", func() {
		complete := batchv1.Job{
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
				},
			},
		}
		Expect(controller.IsJobFinished(&complete)).To(BeTrue())
	})

	It("Returns true for a failed job", func() {
		failed := batchv1.Job{
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{Type: batchv1.JobFailed, Status: corev1.ConditionTrue},
				},
			},
		}
		Expect(controller.IsJobFinished(&failed)).To(BeTrue())
	})

	It("Returns false for a running job", func() {
		running := batchv1.Job{}
		Expect(controller.IsJobFinished(&running)).To(BeFalse())
	})
})
