// Package templatecontroller implements a TemplateReconciler that can
// reconcile Access Templates in a general sense.
package templatecontroller

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

// TemplateReconciler is configured to watch for a particular type
// (TemplateType) of Access Template and then execute the reconciler logic
// against them.
//
// Unlike Access Requests, we don't believe that Templates need significant
// enough validation logic that they warrant their own IBuilder class.
type TemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// APIReader should be generated with mgr.GetAPIReader() to create a non-cached client object.
	// This is used for certain Get() calls where we need to ensure we are getting the latest
	// version from the API, and not a cached object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	APIReader client.Reader

	// TemplateType informs the RequestReconciler what "Kind" of objects it
	// is going to Watch for, and how to retrive them from the Kubernetes API.
	TemplateType v1alpha1.ITemplateResource

	// Builder provides an IBuilder compatible object for handling the RequestType reconciliation
	Builder builders.IBuilder

	// Frequency to re-reconcile successfully reconciled templates
	ReconciliationInterval time.Duration
}

// GetAPIReader conforms to the internal.status.hasStatusReconciler interface.
func (r *TemplateReconciler) GetAPIReader() client.Reader {
	return r.APIReader
}

// RequestContext represents a reconciliation request context.
type RequestContext struct {
	context.Context

	resourceType string
	obj          v1alpha1.ITemplateResource
	req          ctrl.Request
	log          logr.Logger
}

func newRequestContext(
	ctx context.Context,
	sourceObj v1alpha1.ITemplateResource,
	req ctrl.Request,
) *RequestContext {
	// Determine the Resource Type string which will be used for the logger
	resourceType := reflect.TypeOf(sourceObj).String()

	// Create an empty object that we'll be using for the duration of this reconciliation
	emptyObj := sourceObj.DeepCopyObject().(v1alpha1.ITemplateResource)

	return &RequestContext{
		Context:      ctx,
		resourceType: resourceType,
		obj:          emptyObj,
		req:          req,
		log:          ctrl.LoggerFrom(ctx).WithName("TemplateReconciler"),
	}
}
