// Package requestcontroller implements a RequestReconciler that can handle
// Access Requests in a general sense.
package requestcontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RequestReconciler is configured watch for a particular type (RequestType) of
// Access Requests, and execute the reconciler logic against them with a
// particular Builder (Builder). The business logic of what happens in any type
// of Access Request as far as resource creation is all handled inside the
// Builder.
type RequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// RequestType informs the RequestReconciler what "Kind" of objects it
	// is going to Watch for, and how to retrive them from the Kubernetes API.
	RequestType v1alpha1.IRequestResource

	// Builder provides an IBuilder compatible object for handling the RequestType reconciliation
	Builder builders.IBuilder

	// APIReader should be generated with mgr.GetAPIReader() to create a non-cached client object.
	// This is used for certain Get() calls where we need to ensure we are getting the latest
	// version from the API, and not a cached object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	APIReader client.Reader

	// Frequency to re-reconcile successfully reconciled requests
	ReconcilliationInterval time.Duration
}

// GetAPIReader conforms to the internal.status.hasStatusReconciler interface.
func (r *RequestReconciler) GetAPIReader() client.Reader {
	return r.APIReader
}

// RequestContext represents a reconciliation request context.
type RequestContext struct {
	context.Context

	resourceType string
	obj          v1alpha1.IRequestResource
	req          ctrl.Request
	log          logr.Logger
}

func newRequestContext(
	ctx context.Context,
	sourceObj v1alpha1.IRequestResource,
	req ctrl.Request,
) *RequestContext {
	// Determine the Resource Type string which will be used for the logger
	resourceType := reflect.TypeOf(sourceObj).String()

	// Create an empty object that we'll be using for the duration of this reconciliation
	emptyObj := sourceObj.DeepCopyObject().(v1alpha1.IRequestResource)

	return &RequestContext{
		Context:      ctx,
		resourceType: resourceType,
		obj:          emptyObj,
		req:          req,
		log:          ctrl.LoggerFrom(ctx).WithName("RequestReconciler"),
	}
}
