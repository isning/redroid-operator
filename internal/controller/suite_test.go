// Package controller_test contains unit tests for the redroid-operator controllers.
// Tests use controller-runtime's fake client to avoid requiring a real cluster or envtest binaries.
package controller_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

// newTestScheme builds a Scheme that includes all types used by the controllers.
func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := batchv1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := redroidv1alpha1.AddToScheme(s); err != nil {
		panic(err)
	}
	return s
}
