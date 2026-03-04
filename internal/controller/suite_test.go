// Package controller_test contains unit tests for the redroid-operator controllers.
// Tests use controller-runtime's fake client to avoid requiring a real cluster or envtest binaries.
package controller_test

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

// newTestScheme builds a Scheme that includes all types used by the controllers.
func newTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatalf("add clientgo scheme: %v", err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("add corev1 scheme: %v", err)
	}
	if err := batchv1.AddToScheme(s); err != nil {
		t.Fatalf("add batchv1 scheme: %v", err)
	}
	if err := redroidv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add redroid scheme: %v", err)
	}
	return s
}
