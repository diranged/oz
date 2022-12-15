// Package status provides a simple mechanism for updating the Status of an
// v1alpha1.ICoreResource resource
package status

import "sigs.k8s.io/controller-runtime/pkg/client"

// hasStatusReconciler provides an internal interface for a Reconciler that has
// the methods we need for pushing updates to the /status endpoint and
// refetching from a non-cached API Reader.
type hasStatusReconciler interface {
	// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/client/interfaces.go#L82-L86
	Status() client.StatusWriter

	// Returns a read-only API Client without any caching.
	GetAPIReader() client.Reader
}
