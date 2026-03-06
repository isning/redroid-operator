package controller_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

func makeInstance(name string, index int, suspended bool) *redroidv1alpha1.RedroidInstance {
	return &redroidv1alpha1.RedroidInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: redroidv1alpha1.RedroidInstanceSpec{
			Index:         index,
			Image:         "redroid/redroid:16.0.0-latest",
			Suspend:       suspended,
			SharedDataPVC: "redroid-data-base-pvc",
			DiffDataPVC:   "redroid-data-diff-pvc",
			GPUMode:       "host",
		},
	}
}

func reconcileInstance(r *controller.RedroidInstanceReconciler, name string) ctrl.Result {
	res, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
	})
	Expect(err).NotTo(HaveOccurred(), "Reconcile returned error")
	return res
}

var _ = Describe("RedroidInstance Controller", func() {
	var (
		fakeClient client.Client
		r          *controller.RedroidInstanceReconciler
		scheme     = newTestScheme()
	)

	BeforeEach(func() {
	})

	Context("Reconciling a RedroidInstance", func() {
		It("Adds a finalizer on the first reconcile", func() {
			inst := makeInstance("test-0", 0, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "test-0")

			updated := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-0", Namespace: "default"}, updated)
			Expect(err).NotTo(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(updated, "redroid.isning.moe/instance-finalizer")).To(BeTrue(), "expected finalizer to be set after first reconcile")
		})

		It("Creates a Pod for an unsuspended instance", func() {
			inst := makeInstance("inst-active", 2, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-active") // adds finalizer
			reconcileInstance(r, "inst-active") // creates Pod

			fakeRec := r.Recorder.(*record.FakeRecorder)
			Expect(fakeRec.Events).To(Receive(ContainSubstring("CreatedPod")))

			podName := fmt.Sprintf("redroid-instance-%s", "inst-active")
			pod := &corev1.Pod{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod)
			Expect(err).NotTo(HaveOccurred(), "expected Pod to exist")
			Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
			Expect(pod.Spec.InitContainers).To(HaveLen(1)) // kmsg-tools (default-on)
			Expect(pod.Spec.Containers).To(HaveLen(1))     // redroid

			c := pod.Spec.Containers[0]
			Expect(c.Name).To(Equal("redroid"))
			Expect(c.Image).To(Equal("redroid/redroid:16.0.0-latest"))
			Expect(c.ReadinessProbe).NotTo(BeNil())
			Expect(c.ReadinessProbe.Exec.Command).To(ContainElement(ContainSubstring("sys.boot_completed")))
		})

		It("Creates no Pod when suspended=true", func() {
			inst := makeInstance("inst-suspended", 3, true)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-suspended")
			reconcileInstance(r, "inst-suspended")

			podName := fmt.Sprintf("redroid-instance-%s", "inst-suspended")
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{})
			Expect(err).To(HaveOccurred(), "expected no Pod to exist when suspended=true")
		})

		It("Deletes a running Pod when suspended", func() {
			inst := makeInstance("inst-toggle", 1, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-toggle")
			reconcileInstance(r, "inst-toggle")

			podName := fmt.Sprintf("redroid-instance-%s", "inst-toggle")
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{})
			Expect(err).NotTo(HaveOccurred(), "Pod should exist before suspend")

			updated := &redroidv1alpha1.RedroidInstance{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-toggle", Namespace: "default"}, updated)
			Expect(err).NotTo(HaveOccurred())

			updated.Spec.Suspend = true
			err = fakeClient.Update(context.Background(), updated)
			Expect(err).NotTo(HaveOccurred())

			fakeRec := r.Recorder.(*record.FakeRecorder)
			for len(fakeRec.Events) > 0 {
				<-fakeRec.Events
			}

			reconcileInstance(r, "inst-toggle")

			Expect(fakeRec.Events).To(Receive(ContainSubstring("DeletedPod")))

			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, &corev1.Pod{})
			Expect(err).To(HaveOccurred(), "expected Pod to be deleted after suspend")
		})

		It("Sets status to Pending just after Pod creation", func() {
			inst := makeInstance("inst-status", 0, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-status")
			reconcileInstance(r, "inst-status")

			result := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-status", Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstancePending))
			wantPod := fmt.Sprintf("redroid-instance-%s", "inst-status")
			Expect(result.Status.PodName).To(Equal(wantPod))
		})

		It("Sets ADBAddress when Pod is Running", func() {
			inst := makeInstance("inst-running", 4, false)
			podName := fmt.Sprintf("redroid-instance-%s", "inst-running")
			runningPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.42"},
			}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst, runningPod).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-running")
			reconcileInstance(r, "inst-running")

			result := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-running", Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceRunning))
			Expect(result.Status.ADBAddress).To(Equal("10.0.0.42:5555"))
		})

		It("Sets status to Failed when Pod failed", func() {
			inst := makeInstance("inst-failed", 5, false)
			podName := fmt.Sprintf("redroid-instance-%s", "inst-failed")
			failedPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodFailed},
			}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst, failedPod).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-failed")
			reconcileInstance(r, "inst-failed")

			result := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-failed", Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceFailed))
		})

		It("Sets status to Stopped when Pod Succeeded", func() {
			inst := makeInstance("inst-succeeded", 6, false)
			podName := fmt.Sprintf("redroid-instance-%s", "inst-succeeded")
			succeededPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
			}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst, succeededPod).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-succeeded")
			reconcileInstance(r, "inst-succeeded")

			result := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-succeeded", Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceStopped))
		})

		It("Mounts overlayfs args properly in Pod start command", func() {
			inst := makeInstance("inst-args", 7, false)
			inst.Spec.ExtraArgs = []string{"androidboot.redroid_width=1080"}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-args")
			reconcileInstance(r, "inst-args")

			podName := fmt.Sprintf("redroid-instance-%s", "inst-args")
			pod := &corev1.Pod{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod)
			Expect(err).NotTo(HaveOccurred())

			wantArgs := map[string]bool{
				"androidboot.redroid_gpu_mode=host":   true,
				"androidboot.use_memfd=1":             true,
				"androidboot.use_redroid_overlayfs=1": true,
				"androidboot.redroid_width=1080":      true,
			}
			for _, a := range pod.Spec.Containers[0].Args {
				delete(wantArgs, a)
			}
			Expect(wantArgs).To(BeEmpty(), "missing expected args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
		})

		It("Applies specific volume mounts to the Pod (/data-base, /data-diff, /dev/dri)", func() {
			inst := makeInstance("inst-vols", 3, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "inst-vols")
			reconcileInstance(r, "inst-vols")

			podName := fmt.Sprintf("redroid-instance-%s", "inst-vols")
			pod := &corev1.Pod{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod)
			Expect(err).NotTo(HaveOccurred())

			mountPaths := map[string]bool{}
			for _, vm := range pod.Spec.Containers[0].VolumeMounts {
				mountPaths[vm.MountPath] = true
			}
			for _, want := range []string{"/data-base", "/data-diff/3", "/dev/dri"} {
				Expect(mountPaths[want]).To(BeTrue(), "expected mount path %q not found; mounts: %v", want, mountPaths)
			}
		})

		It("Returns no error for missing objects", func() {
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			_, err := r.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "does-not-exist", Namespace: "default"},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Cleans up the Pod when the Instance is deleted", func() {
			inst := makeInstance("del-0", 0, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}

			reconcileInstance(r, "del-0") // adds finalizer
			reconcileInstance(r, "del-0") // creates Pod

			podList := &corev1.PodList{}
			err := fakeClient.List(context.Background(), podList)
			Expect(err).NotTo(HaveOccurred())
			Expect(podList.Items).NotTo(BeEmpty())

			current := &redroidv1alpha1.RedroidInstance{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "del-0", Namespace: "default"}, current)
			Expect(err).NotTo(HaveOccurred())

			err = fakeClient.Delete(context.Background(), current)
			Expect(err).NotTo(HaveOccurred())

			reconcileInstance(r, "del-0")

			final := &redroidv1alpha1.RedroidInstance{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "del-0", Namespace: "default"}, final)
			if err == nil {
				Expect(controllerutil.ContainsFinalizer(final, "redroid.isning.moe/instance-finalizer")).To(BeFalse())
			}
		})

		It("Sets status to Stopped when Pod Succeeded (handles process exit/shutdown)", func() {
			inst := makeInstance("succ-0", 0, false)
			podName := "redroid-instance-succ-0"
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
			}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst, pod).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "succ-0") // adds finalizer
			reconcileInstance(r, "succ-0") // observes Succeeded Pod

			result := &redroidv1alpha1.RedroidInstance{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "succ-0", Namespace: "default"}, result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceStopped))
		})

		It("Creates a stable ClusterIP Service for the instance", func() {
			inst := makeInstance("svc-test", 0, false)

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "svc-test") // adds finalizer
			reconcileInstance(r, "svc-test") // creates Pod + Service

			svcName := "redroid-instance-svc-test"
			svc := &corev1.Service{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: svcName, Namespace: "default"}, svc)
			Expect(err).NotTo(HaveOccurred())

			Expect(svc.Spec.Selector["redroid.isning.moe/instance"]).To(Equal("svc-test"))
			Expect(svc.Spec.Ports).To(HaveLen(1))
			Expect(svc.Spec.Ports[0].Port).To(Equal(int32(5555)))
			Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
		})

		It("Creates the Service even if suspended", func() {
			inst := makeInstance("svc-suspended", 0, true) // suspended from the start

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "svc-suspended") // adds finalizer
			reconcileInstance(r, "svc-suspended") // reconciles suspended path

			svcName := "redroid-instance-svc-suspended"
			svc := &corev1.Service{}
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: svcName, Namespace: "default"}, svc)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Applies custom Service configuration (spec.service fields)", func() {
			nodePort := int32(30555)
			inst := makeInstance("svc-custom", 0, false)
			inst.Spec.Service = &redroidv1alpha1.InstanceServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				NodePort: &nodePort,
				Annotations: map[string]string{
					"example.io/custom": "true",
				},
			}

			fakeClient = fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
				WithObjects(inst).Build()

			r = &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
			reconcileInstance(r, "svc-custom") // adds finalizer
			reconcileInstance(r, "svc-custom") // creates Service

			svc := &corev1.Service{}
			err := fakeClient.Get(context.Background(),
				types.NamespacedName{Name: "redroid-instance-svc-custom", Namespace: "default"}, svc)
			Expect(err).NotTo(HaveOccurred())

			Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			Expect(svc.Spec.Ports).To(HaveLen(1))
			Expect(svc.Spec.Ports[0].NodePort).To(Equal(nodePort))
			Expect(svc.Annotations["example.io/custom"]).To(Equal("true"))
		})
	})
})
