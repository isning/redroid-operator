package controller_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
	"github.com/isning/redroid-operator/internal/controller"
)

// TestRedroidInstance_FPSArg verifies androidboot.redroid_fps is passed when screen.fps is set.
func TestRedroidInstance_FPSArg(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-fps", 0, false)
	fps := int32(60)
	inst.Spec.Screen = &redroidv1alpha1.ScreenSpec{FPS: &fps}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-fps")
	reconcileInstance(t, r, "inst-fps")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-fps", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	want := "androidboot.redroid_fps=60"
	for _, a := range pod.Spec.Containers[0].Args {
		if a == want {
			return
		}
	}
	t.Errorf("expected arg %q not found in %v", want, pod.Spec.Containers[0].Args)
}

// TestRedroidInstance_GPUNode verifies androidboot.redroid_gpu_node is passed when gpuNode is set.
func TestRedroidInstance_GPUNode(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-gpunode", 0, false)
	inst.Spec.GPUNode = "/dev/dri/renderD128"

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-gpunode")
	reconcileInstance(t, r, "inst-gpunode")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-gpunode", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	want := "androidboot.redroid_gpu_node=/dev/dri/renderD128"
	for _, a := range pod.Spec.Containers[0].Args {
		if a == want {
			return
		}
	}
	t.Errorf("expected arg %q not found in %v", want, pod.Spec.Containers[0].Args)
}

// TestRedroidInstance_NetworkDNS verifies DNS args (net_ndns, net_dns1, net_dns2) are generated.
func TestRedroidInstance_NetworkDNS(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-dns", 0, false)
	inst.Spec.Network = &redroidv1alpha1.NetworkSpec{
		DNS: []string{"8.8.8.8", "8.8.4.4"},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-dns")
	reconcileInstance(t, r, "inst-dns")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-dns", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	wantArgs := map[string]bool{
		"androidboot.redroid_net_ndns=2":       true,
		"androidboot.redroid_net_dns1=8.8.8.8": true,
		"androidboot.redroid_net_dns2=8.8.4.4": true,
	}
	for _, a := range pod.Spec.Containers[0].Args {
		delete(wantArgs, a)
	}
	if len(wantArgs) > 0 {
		t.Errorf("missing DNS args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	}
}

// TestRedroidInstance_NetworkProxy verifies proxy args are generated from NetworkSpec.proxy.
func TestRedroidInstance_NetworkProxy(t *testing.T) {
	scheme := newTestScheme(t)
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

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-proxy")
	reconcileInstance(t, r, "inst-proxy")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-proxy", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	wantArgs := map[string]bool{
		"androidboot.redroid_net_proxy_type=static":                      true,
		"androidboot.redroid_net_proxy_host=proxy.example.com":           true,
		"androidboot.redroid_net_proxy_port=8080":                        true,
		"androidboot.redroid_net_proxy_exclude_list=localhost,127.0.0.1": true,
	}
	for _, a := range pod.Spec.Containers[0].Args {
		delete(wantArgs, a)
	}
	if len(wantArgs) > 0 {
		t.Errorf("missing proxy args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	}
}

// TestRedroidInstance_NetworkProxyPAC verifies PAC proxy args.
func TestRedroidInstance_NetworkProxyPAC(t *testing.T) {
	scheme := newTestScheme(t)
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

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-pac")
	reconcileInstance(t, r, "inst-pac")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-pac", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	wantArgs := map[string]bool{
		"androidboot.redroid_net_proxy_type=pac":                               true,
		"androidboot.redroid_net_proxy_pac=http://proxy.example.com/proxy.pac": true,
	}
	for _, a := range pod.Spec.Containers[0].Args {
		delete(wantArgs, a)
	}
	if len(wantArgs) > 0 {
		t.Errorf("missing PAC proxy args: %v; got: %v", wantArgs, pod.Spec.Containers[0].Args)
	}
}

// TestRedroidInstance_ExtraEnvPassedToContainer verifies ExtraEnv entries appear in the container Env.
func TestRedroidInstance_ExtraEnvPassedToContainer(t *testing.T) {
	scheme := newTestScheme(t)
	inst := makeInstance("inst-extraenv", 0, false)
	inst.Spec.ExtraEnv = []corev1.EnvVar{
		{Name: "MY_CUSTOM_VAR", Value: "hello"},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&redroidv1alpha1.RedroidInstance{}).
		WithObjects(inst).Build()

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-extraenv")
	reconcileInstance(t, r, "inst-extraenv")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-extraenv", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	for _, e := range pod.Spec.Containers[0].Env {
		if e.Name == "MY_CUSTOM_VAR" && e.Value == "hello" {
			return
		}
	}
	t.Errorf("env MY_CUSTOM_VAR=hello not found in %v", pod.Spec.Containers[0].Env)
}

// TestRedroidInstance_ExtraEnvSecretRef verifies ExtraEnv with secretKeyRef is propagated to the container.
func TestRedroidInstance_ExtraEnvSecretRef(t *testing.T) {
	scheme := newTestScheme(t)
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

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-secretenv")
	reconcileInstance(t, r, "inst-secretenv")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-secretenv", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	// Verify env var with secretKeyRef is present.
	var found bool
	for _, e := range pod.Spec.Containers[0].Env {
		if e.Name == "PROXY_HOST" && e.ValueFrom != nil &&
			e.ValueFrom.SecretKeyRef != nil &&
			e.ValueFrom.SecretKeyRef.Name == "proxy-secret" &&
			e.ValueFrom.SecretKeyRef.Key == "host" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected PROXY_HOST secretKeyRef env in %v", pod.Spec.Containers[0].Env)
	}

	// Verify the extra arg referencing the env var is present.
	wantArg := "androidboot.redroid_net_proxy_host=$(PROXY_HOST)"
	for _, a := range pod.Spec.Containers[0].Args {
		if a == wantArg {
			return
		}
	}
	t.Errorf("expected arg %q not found in %v", wantArg, pod.Spec.Containers[0].Args)
}

// TestRedroidInstance_ExtraEnvConfigMapRef verifies ExtraEnv with configMapKeyRef is propagated.
func TestRedroidInstance_ExtraEnvConfigMapRef(t *testing.T) {
	scheme := newTestScheme(t)
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

	r := &controller.RedroidInstanceReconciler{Client: fakeClient, Scheme: scheme}
	reconcileInstance(t, r, "inst-cmenv")
	reconcileInstance(t, r, "inst-cmenv")

	pod := &corev1.Pod{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "redroid-instance-inst-cmenv", Namespace: "default"}, pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}
	for _, e := range pod.Spec.Containers[0].Env {
		if e.Name == "REDROID_WIDTH_OVERRIDE" &&
			e.ValueFrom != nil &&
			e.ValueFrom.ConfigMapKeyRef != nil &&
			e.ValueFrom.ConfigMapKeyRef.Name == "redroid-config" &&
			e.ValueFrom.ConfigMapKeyRef.Key == "width" {
			return
		}
	}
	t.Errorf("expected REDROID_WIDTH_OVERRIDE configMapKeyRef env in %v", pod.Spec.Containers[0].Env)
}

// ---- Unused import guard ----
var _ = metav1.Now
