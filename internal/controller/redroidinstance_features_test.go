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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

var _ = Describe("RedroidInstance Features", func() {
	var (
		scheme = newTestScheme()
	)

	It("Passes screen resolution args to the redroid container", func() {
		inst := makeInstance("inst-screen", 0, false)
		width, height, dpi := int32(1080), int32(1920), int32(480)
		inst.Spec.Screen = &redroidv1alpha1.ScreenSpec{Width: &width, Height: &height, DPI: &dpi}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-screen")
		reconcileInstance(r, "inst-screen")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-screen", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		wantArgs := map[string]bool{
			"androidboot.redroid_width=1080":  true,
			"androidboot.redroid_height=1920": true,
			"androidboot.redroid_dpi=480":     true,
		}
		for _, a := range pod.Spec.Containers[0].Args {
			delete(wantArgs, a)
		}
		Expect(wantArgs).To(BeEmpty(), "missing screen args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	})

	It("Applies custom ADB port to Pod and Status", func() {
		inst := makeInstance("inst-port", 0, false)
		customPort := int32(6666)
		inst.Spec.ADBPort = &customPort

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-port") // adds finalizer
		reconcileInstance(r, "inst-port") // creates Pod

		// Check the controller-created Pod uses the custom port.
		podName := "redroid-instance-inst-port"
		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		Expect(pod.Spec.Containers).NotTo(BeEmpty(), "expected at least one container")

		found := false
		for _, p := range pod.Spec.Containers[0].Ports {
			if p.ContainerPort == customPort {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "expected container port %d, got: %v", customPort, pod.Spec.Containers[0].Ports)

		// Simulate pod going Running, check ADBAddress uses the custom port.
		pod.Status.Phase = corev1.PodRunning
		pod.Status.PodIP = "10.1.0.5"
		err = fakeClient.Status().Update(context.Background(), pod)
		Expect(err).NotTo(HaveOccurred(), "update pod status")

		reconcileInstance(r, "inst-port")

		result := &redroidv1alpha1.RedroidInstance{}
		err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-port", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		want := fmt.Sprintf("10.1.0.5:%d", customPort)
		Expect(result.Status.ADBAddress).To(Equal(want))
	})

	It("Sets Ready condition to True when Pod is Running", func() {
		inst := makeInstance("inst-cond", 0, false)
		podName := "redroid-instance-inst-cond"
		runningPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.99"},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst, runningPod).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-cond")
		reconcileInstance(r, "inst-cond")

		result := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-cond", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		found := false
		for _, c := range result.Status.Conditions {
			if c.Type == string(redroidv1alpha1.RedroidInstanceConditionReady) {
				found = true
				Expect(c.Status).To(Equal(metav1.ConditionTrue), "expected Ready=True")
			}
		}
		Expect(found).To(BeTrue(), "Ready condition not found in: %v", result.Status.Conditions)
	})

	It("Sets ObservedGeneration in Status", func() {
		inst := makeInstance("inst-gen", 0, false)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-gen")
		reconcileInstance(r, "inst-gen")

		result := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-gen", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		Expect(result.Status.ObservedGeneration).To(Equal(result.Generation))
	})

	It("Propagates tolerations to the Pod spec", func() {
		inst := makeInstance("inst-tol", 0, false)
		inst.Spec.Tolerations = []corev1.Toleration{
			{Key: "gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-tol")
		reconcileInstance(r, "inst-tol")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-tol", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		Expect(pod.Spec.Tolerations).NotTo(BeEmpty(), "expected tolerations to be set on Pod")
		Expect(pod.Spec.Tolerations[0].Key).To(Equal("gpu"))
	})

	It("Sets Scheduled=False when suspended", func() {
		inst := makeInstance("inst-susp-cond", 0, true)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-susp-cond")

		result := &redroidv1alpha1.RedroidInstance{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "inst-susp-cond", Namespace: "default"}, result)
		Expect(err).NotTo(HaveOccurred(), "get instance")

		Expect(result.Status.Phase).To(Equal(redroidv1alpha1.RedroidInstanceStopped))

		for _, c := range result.Status.Conditions {
			if c.Type == string(redroidv1alpha1.RedroidInstanceConditionReady) {
				Expect(c.Status).NotTo(Equal(metav1.ConditionTrue), "Ready condition should not be True when suspended")
			}
		}
	})

	It("Mounts sharedDataPVC directly as /data in BaseMode and disables overlayfs", func() {
		inst := makeInstance("inst-base", 0, false)
		inst.Spec.BaseMode = true

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-base")
		reconcileInstance(r, "inst-base")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-base", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		// Overlayfs must be disabled in base mode.
		foundOverlayfsOff := false
		for _, a := range pod.Spec.Containers[0].Args {
			if a == "androidboot.use_redroid_overlayfs=0" {
				foundOverlayfsOff = true
			}
			Expect(a).NotTo(Equal("androidboot.use_redroid_overlayfs=1"), "overlayfs must be 0 in baseMode, got 1")
		}
		Expect(foundOverlayfsOff).To(BeTrue(), "expected androidboot.use_redroid_overlayfs=0 in args; got: %v", pod.Spec.Containers[0].Args)

		// /data must be mounted from sharedDataPVC.
		foundDataMount := false
		for _, vm := range pod.Spec.Containers[0].VolumeMounts {
			if vm.MountPath == "/data" {
				foundDataMount = true
			}
			Expect(vm.MountPath).NotTo(Equal("/data-base"), "base mode must not mount overlayfs paths")
			Expect(vm.MountPath).NotTo(Equal("/data-diff/0"), "base mode must not mount overlayfs paths")
		}
		Expect(foundDataMount).To(BeTrue(), "expected /data mount in base mode")

		// diff volume must not be present.
		for _, v := range pod.Spec.Volumes {
			Expect(v.Name).NotTo(Equal("data-diff"), "base mode must not include data-diff volume")
		}
	})

	It("Keeps overlayfs enabled and mounts /data-base and /data-diff in NormalMode", func() {
		inst := makeInstance("inst-normal", 0, false) // baseMode defaults to false

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-normal")
		reconcileInstance(r, "inst-normal")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-normal", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		// Overlayfs must be enabled in normal mode.
		foundOverlayfsOn := false
		for _, a := range pod.Spec.Containers[0].Args {
			if a == "androidboot.use_redroid_overlayfs=1" {
				foundOverlayfsOn = true
			}
		}
		Expect(foundOverlayfsOn).To(BeTrue(), "expected androidboot.use_redroid_overlayfs=1, got: %v", pod.Spec.Containers[0].Args)

		// /data-base and data-diff must be present.
		mounts := map[string]bool{}
		for _, vm := range pod.Spec.Containers[0].VolumeMounts {
			mounts[vm.MountPath] = true
		}
		Expect(mounts["/data-base"]).To(BeTrue(), "expected /data-base mount in normal mode")
		Expect(mounts["/data-diff/0"]).To(BeTrue(), "expected /data-diff/0 mount in normal mode")
	})

	It("Adds kmsg-tools init container and wraps main command (default enabled)", func() {
		inst := makeInstance("inst-kmsg", 0, false)
		// DisableKmsgRedirect is false by default

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-kmsg")
		reconcileInstance(r, "inst-kmsg")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-kmsg", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		// Pod must have 1 init container (kmsg-tools) and 1 main container (redroid).
		Expect(pod.Spec.InitContainers).To(HaveLen(1), "expected kmsg-tools init container")
		Expect(pod.Spec.Containers).To(HaveLen(1), "expected only the main redroid container")

		// shareProcessNamespace must be nil (no longer needed).
		Expect(pod.Spec.ShareProcessNamespace).To(BeNil())

		// Init container: copies socat to emptyDir.
		ic := pod.Spec.InitContainers[0]
		Expect(ic.Name).To(Equal("kmsg-tools"))
		Expect(ic.Image).To(Equal("ghcr.io/isning/redroid-operator/kmsg-tools:latest"))
		Expect(ic.Command).To(HaveLen(3))
		Expect(ic.Command[2]).To(ContainSubstring("cp /bin/socat /bin/busybox /kmsg-tools/"))

		// Main container: /kmsg-tools/busybox sh wrapper using injected socat.
		main := pod.Spec.Containers[0]
		Expect(main.Name).To(Equal("redroid"))
		Expect(main.Command).To(Equal([]string{"/kmsg-tools/busybox", "sh"}), "main container must use /kmsg-tools/busybox sh wrapper")
		Expect(main.Args[0]).To(Equal("-c"))
		Expect(main.Args[1]).To(ContainSubstring("/kmsg-tools/socat PTY"), "wrapper must use injected socat")
		Expect(main.Args[1]).To(ContainSubstring("mount --bind /tmp/kmsg-pty /dev/kmsg"), "wrapper must bind-mount PTY over /dev/kmsg")
		Expect(main.Args[1]).To(ContainSubstring(`exec /init "$@"`), "wrapper must exec /init with original args")
		Expect(main.Args[2]).To(Equal("--"))
		// Original androidboot args appear after --.
		Expect(main.Args[3:]).To(ContainElement("androidboot.redroid_gpu_mode=host"))

		// kmsg-tools emptyDir must be mounted in both containers.
		var foundInitMount, foundMainMount bool
		for _, vm := range ic.VolumeMounts {
			if vm.Name == "kmsg-tools" && vm.MountPath == "/kmsg-tools" {
				foundInitMount = true
			}
		}
		for _, vm := range main.VolumeMounts {
			if vm.Name == "kmsg-tools" && vm.MountPath == "/kmsg-tools" {
				foundMainMount = true
			}
		}
		Expect(foundInitMount).To(BeTrue(), "init container must mount kmsg-tools shared volume")
		Expect(foundMainMount).To(BeTrue(), "main container must mount kmsg-tools shared volume")

		// kmsg-tools volume must be declared in the Pod as emptyDir.
		var foundToolsVolume bool
		for _, v := range pod.Spec.Volumes {
			if v.Name == "kmsg-tools" && v.EmptyDir != nil {
				foundToolsVolume = true
			}
		}
		Expect(foundToolsVolume).To(BeTrue(), "pod must have kmsg-tools emptyDir volume")
	})

	It("Disables kmsg inject when DisableKmsgRedirect is true", func() {
		inst := makeInstance("inst-no-kmsg", 0, false)
		inst.Spec.DisableKmsgRedirect = true

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-no-kmsg")
		reconcileInstance(r, "inst-no-kmsg")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-no-kmsg", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		// Must have exactly one container (redroid) and zero init containers.
		Expect(pod.Spec.InitContainers).To(BeEmpty(), "DisableKmsgRedirect=true must suppress the init container")
		Expect(pod.Spec.Containers).To(HaveLen(1))

		// Main container must use the image's default ENTRYPOINT (no wrapper).
		c := pod.Spec.Containers[0]
		Expect(c.Command).To(BeNil(), "command must be nil when DisableKmsgRedirect=true")
		Expect(c.Args).NotTo(ContainElement("-c"), "args must not contain -c shell flag")

		// No kmsg-tools volume.
		for _, v := range pod.Spec.Volumes {
			Expect(v.Name).NotTo(Equal("kmsg-tools"), "kmsg-tools volume must not exist when redirect is disabled")
		}
	})

	It("Uses custom init container image from spec.kmsgToolsImage", func() {
		inst := makeInstance("inst-custom-tools", 0, false)
		inst.Spec.KmsgToolsImage = "my-registry/kmsg-tools:v1.2.3"

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-custom-tools")
		reconcileInstance(r, "inst-custom-tools")

		pod := &corev1.Pod{}
		Expect(fakeClient.Get(context.Background(),
			types.NamespacedName{Name: "redroid-instance-inst-custom-tools", Namespace: "default"}, pod)).
			To(Succeed())

		ic := pod.Spec.InitContainers[0]
		Expect(ic.Name).To(Equal("kmsg-tools"))
		Expect(ic.Image).To(Equal("my-registry/kmsg-tools:v1.2.3"), "custom tools image must be used")
	})

	It("Forwards ExtraArgs through the main container wrapper", func() {
		inst := makeInstance("inst-extra-args", 0, false)
		inst.Spec.ExtraArgs = []string{
			"androidboot.redroid_width=1080",
			"androidboot.redroid_height=1920",
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-extra-args")
		reconcileInstance(r, "inst-extra-args")

		pod := &corev1.Pod{}
		Expect(fakeClient.Get(context.Background(),
			types.NamespacedName{Name: "redroid-instance-inst-extra-args", Namespace: "default"}, pod)).
			To(Succeed())

		mainArgs := pod.Spec.Containers[0].Args
		// Args after "--": the full androidboot arg list forwarded to /init via $@
		Expect(mainArgs[2]).To(Equal("--"))
		Expect(mainArgs[3:]).To(ContainElement("androidboot.redroid_width=1080"))
		Expect(mainArgs[3:]).To(ContainElement("androidboot.redroid_height=1920"))
	})

	It("Adds kmsg-tools emptyDir in BaseMode as well", func() {
		inst := makeInstance("inst-base-kmsg", 0, false)
		inst.Spec.BaseMode = true

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-base-kmsg")
		reconcileInstance(r, "inst-base-kmsg")

		pod := &corev1.Pod{}
		Expect(fakeClient.Get(context.Background(),
			types.NamespacedName{Name: "redroid-instance-inst-base-kmsg", Namespace: "default"}, pod)).
			To(Succeed())

		// EmptyDir volume must exist in BaseMode too.
		var found bool
		for _, v := range pod.Spec.Volumes {
			if v.Name == "kmsg-tools" && v.EmptyDir != nil {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "kmsg-tools emptyDir must also be added in BaseMode")

		// Main container must still carry the sync VolumeMount.
		var foundMount bool
		for _, vm := range pod.Spec.Containers[0].VolumeMounts {
			if vm.Name == "kmsg-tools" {
				foundMount = true
			}
		}
		Expect(foundMount).To(BeTrue(), "main container must mount kmsg-tools in BaseMode")
	})
})
