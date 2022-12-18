package controllers

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BaseReconciler extends the default reconciler behaviors (client.Client+Scheme) and provide some
// helper functions for refetching objects directly from the API, pushing status updates, etc.
type BaseReconciler struct {
	// Extend the standard client.Client interface, which is a requirement for the base
	// reconciliation code
	client.Client
	Scheme *runtime.Scheme

	// APIReader should be generated with mgr.GetAPIReader() to create a non-cached client object.
	// This is used for certain Get() calls where we need to ensure we are getting the latest
	// version from the API, and not a cached object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	APIReader client.Reader

	// Storage of our logger object - so we don't have to keep getting it from the context. Set by the
	// GetLogger() method.
	logger logr.Logger

	// ReconciliationInterval is the time to wait inbetween re-reconciling ExecAccessRequests. This
	// primarily matters for setting the maximum time after an AccessRequest has expired that it
	// will be purged by the controller.
	ReconcililationInterval int
}

// GetAPIReader conforms to the internal.status.hasStatusReconciler interface.
func (r *BaseReconciler) GetAPIReader() client.Reader {
	return r.APIReader
}

// SetReconciliationInterval sets the BaseReconciler.ReconciliationInterval value to the
// DEFAULT_RECONCILIATION_INTERVAL if it was not pre-populated.
func (r *BaseReconciler) SetReconciliationInterval() {
	if r.ReconcililationInterval == 0 {
		r.ReconcililationInterval = DefaultReconciliationInterval
	}
}
