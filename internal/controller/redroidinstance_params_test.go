package controller_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

var _ = Describe("RedroidInstance Params", func() {
	var (
		scheme = newTestScheme()
	)

	It("Passes androidboot.redroid_fps when screen.fps is set", func() {
		inst := makeInstance("inst-fps", 0, false)
		fps := int32(60)
		inst.Spec.Screen = &redroidv1alpha1.ScreenSpec{FPS: &fps}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-fps")
		reconcileInstance(r, "inst-fps")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-fps", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		want := "androidboot.redroid_fps=60"
		Expect(pod.Spec.Containers[0].Args).To(ContainElement(want))
	})

	It("Passes androidboot.redroid_gpu_node when gpuNode is set", func() {
		inst := makeInstance("inst-gpunode", 0, false)
		inst.Spec.GPUNode = "/dev/dri/renderD128"

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-gpunode")
		reconcileInstance(r, "inst-gpunode")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-gpunode", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		want := "androidboot.redroid_gpu_node=/dev/dri/renderD128"
		Expect(pod.Spec.Containers[0].Args).To(ContainElement(want))
	})

	It("Generates DNS args (net_ndns, net_dns1, net_dns2)", func() {
		inst := makeInstance("inst-dns", 0, false)
		inst.Spec.Network = &redroidv1alpha1.NetworkSpec{
			DNS: []string{"8.8.8.8", "8.8.4.4"},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-dns")
		reconcileInstance(r, "inst-dns")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-dns", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		wantArgs := map[string]bool{
			"androidboot.redroid_net_ndns=2":       true,
			"androidboot.redroid_net_dns1=8.8.8.8": true,
			"androidboot.redroid_net_dns2=8.8.4.4": true,
		}
		for _, a := range pod.Spec.Containers[0].Args {
			delete(wantArgs, a)
		}
		Expect(wantArgs).To(BeEmpty(), "missing DNS args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	})

	It("Generates proxy args from NetworkSpec.proxy", func() {
		inst := makeInstance("inst-proxy", 0, false)
		proxyPort := int32(8080)
		inst.Spec.Network = &redroidv1alpha1.NetworkSpec{
			Proxy: &redroidv1alpha1.ProxySpec{
				Type:        "static",
				Host:        "proxy.example.com",
				Port:        &proxyPort,
				ExcludeList: "localhost,127.0.0.1",
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-proxy")
		reconcileInstance(r, "inst-proxy")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-proxy", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		wantArgs := map[string]bool{
			"androidboot.redroid_net_proxy_type=static":                      true,
			"androidboot.redroid_net_proxy_host=proxy.example.com":           true,
			"androidboot.redroid_net_proxy_port=8080":                        true,
			"androidboot.redroid_net_proxy_exclude_list=localhost,127.0.0.1": true,
		}
		for _, a := range pod.Spec.Containers[0].Args {
			delete(wantArgs, a)
		}
		Expect(wantArgs).To(BeEmpty(), "missing proxy args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	})

	It("Generates PAC proxy args", func() {
		inst := makeInstance("inst-pac", 0, false)
		inst.Spec.Network = &redroidv1alpha1.NetworkSpec{
			Proxy: &redroidv1alpha1.ProxySpec{
				Type: "pac",
				PAC:  "http://proxy.example.com/proxy.pac",
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-pac")
		reconcileInstance(r, "inst-pac")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-pac", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		wantArgs := map[string]bool{
			"androidboot.redroid_net_proxy_type=pac":                               true,
			"androidboot.redroid_net_proxy_pac=http://proxy.example.com/proxy.pac": true,
		}
		for _, a := range pod.Spec.Containers[0].Args {
			delete(wantArgs, a)
		}
		Expect(wantArgs).To(BeEmpty(), "missing PAC proxy args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	})

	It("Passes ExtraEnv to the container Env", func() {
		inst := makeInstance("inst-extraenv", 0, false)
		inst.Spec.ExtraEnv = []corev1.EnvVar{
			{Name: "MY_CUSTOM_VAR", Value: "hello"},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-extraenv")
		reconcileInstance(r, "inst-extraenv")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-extraenv", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		found := false
		for _, e := range pod.Spec.Containers[0].Env {
			if e.Name == "MY_CUSTOM_VAR" && e.Value == "hello" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "env MY_CUSTOM_VAR=hello not found in %v", pod.Spec.Containers[0].Env)
	})

	It("Propagates ExtraEnv with secretKeyRef to the container", func() {
		inst := makeInstance("inst-secretenv", 0, false)
		inst.Spec.ExtraEnv = []corev1.EnvVar{
			{
				Name: "PROXY_HOST",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "proxy-secret"},
						Key:                  "host",
					},
				},
			},
		}
		inst.Spec.ExtraArgs = []string{"androidboot.redroid_net_proxy_host=$(PROXY_HOST)"}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-secretenv")
		reconcileInstance(r, "inst-secretenv")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-secretenv", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		// Verify env var with secretKeyRef is present.
		found := false
		for _, e := range pod.Spec.Containers[0].Env {
			if e.Name == "PROXY_HOST" && e.ValueFrom != nil &&
				e.ValueFrom.SecretKeyRef != nil &&
				e.ValueFrom.SecretKeyRef.Name == "proxy-secret" &&
				e.ValueFrom.SecretKeyRef.Key == "host" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "expected PROXY_HOST secretKeyRef env in %v", pod.Spec.Containers[0].Env)

		// Verify the extra arg referencing the env var is present.
		wantArg := "androidboot.redroid_net_proxy_host=$(PROXY_HOST)"
		Expect(pod.Spec.Containers[0].Args).To(ContainElement(wantArg))
	})

	It("Propagates ExtraEnv with configMapKeyRef to the container", func() {
		inst := makeInstance("inst-cmenv", 0, false)
		inst.Spec.ExtraEnv = []corev1.EnvVar{
			{
				Name: "REDROID_WIDTH_OVERRIDE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "redroid-config"},
						Key:                  "width",
					},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
			WithObjects(inst).Build()

		r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme, Recorder: record.NewFakeRecorder(100)}
		reconcileInstance(r, "inst-cmenv")
		reconcileInstance(r, "inst-cmenv")

		pod := &corev1.Pod{}
		err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-cmenv", Namespace: "default"}, pod)
		Expect(err).NotTo(HaveOccurred(), "get pod")

		found := false
		for _, e := range pod.Spec.Containers[0].Env {
			if e.Name == "REDROID_WIDTH_OVERRIDE" &&
				e.ValueFrom != nil &&
				e.ValueFrom.ConfigMapKeyRef != nil &&
				e.ValueFrom.ConfigMapKeyRef.Name == "redroid-config" &&
				e.ValueFrom.ConfigMapKeyRef.Key == "width" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "expected REDROID_WIDTH_OVERRIDE configMapKeyRef env in %v", pod.Spec.Containers[0].Env)
	})
})
