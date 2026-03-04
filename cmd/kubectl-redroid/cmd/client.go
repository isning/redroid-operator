package cmd

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

// instancesGVR is the GVR for RedroidInstance.
var instancesGVR = schema.GroupVersionResource{Group: "redroid.io", Version: "v1alpha1", Resource: "redroidinstances"}

// tasksGVR is the GVR for RedroidTask.
var tasksGVR = schema.GroupVersionResource{Group: "redroid.io", Version: "v1alpha1", Resource: "redroidtasks"}

// clients bundles the various k8s clients needed by the plugin.
type clients struct {
	cfg     *rest.Config
	kube    kubernetes.Interface
	dynamic dynamic.Interface
	scheme  *runtime.Scheme
}

// buildClients constructs all k8s clients from the current kubeconfig.
func buildClients() (*clients, error) {
	cfg, err := buildClientConfig().ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build REST config: %w", err)
	}

	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build kube client: %w", err)
	}

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build dynamic client: %w", err)
	}

	scheme := runtime.NewScheme()
	_ = redroidv1alpha1.AddToScheme(scheme)

	return &clients{cfg: cfg, kube: kube, dynamic: dyn, scheme: scheme}, nil
}
