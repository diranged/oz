package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// This interface wraps the standard client.Object resource (metav1.Object + runtime.Object) with a requirement for
// a `GetConditions()` function that returns back the nested Status.Conditions list. This is used by
// BaseReconciler.UpdateCondition()
type clientObjectWithConditions interface {
	metav1.Object
	runtime.Object

	// Returns a pointer to a list of conditions. The pointer is important so that the returned value can be
	// updated and then the resource can be saved with the updated conditions.
	GetConditions() *[]metav1.Condition
}

type OzReconciler interface {
	GetLogger(ctx context.Context) logr.Logger
	Refetch(ctx context.Context, obj client.Object) *client.Object
}

type BaseReconciler struct {
	OzReconciler

	// Extend the standard client.Client interface, which is a requirement for the base reconciliation code
	client.Client
	Scheme *runtime.Scheme

	// ApiReader should be generated with mgr.GetAPIReader() to create a non-cached client object. This is used
	// for certain Get() calls where we need to ensure we are getting the latest version from the API, and not a cached
	// object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	ApiReader client.Reader

	// Storage of our logger object - so we don't have to keep getting it from the context. Set by the
	// GetLogger() method.
	logger logr.Logger
}

// Refetch uses the "consistent client" (non-caching) to retreive the latest state of the object into the
// supplied object reference. This is critical to avoid "the object has been modified; please apply
// your changes to the latest version and try again" errors when updating object status fields.
func (b *BaseReconciler) Refetch(ctx context.Context, obj client.Object) error {
	err := b.ApiReader.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, obj)
	return err
}

func (b *BaseReconciler) UpdateStatus(ctx context.Context, obj client.Object) error {
	logger := b.GetLogger(ctx)

	// Update the status, handle failure.
	if err := b.Status().Update(ctx, obj); err != nil {
		logger.Error(err, fmt.Sprintf("Failed to update %s status", obj.GetObjectKind().GroupVersionKind().Kind))
		return err
	}

	// Refetch the object when we're done to make sure we are working with the latest version
	b.Refetch(ctx, obj)

	return nil
}

// UpdateCondition provides a simple way to update the .Status.Conditions field of a given resource. The resource
// must match the ResourceWithConditions interface - which exposes the GetConditions() method.
//
// When an UpdateCondition() call is made, we retrieve the current list of conditions first from the request object.
// From there, we insert in a new Condition into the resource.
// Finally we call the UpdateStatus() function to push the update to Kubernetes.
func (r *BaseReconciler) UpdateCondition(
	ctx context.Context,
	res clientObjectWithConditions,
	conditionType BaseResourceConditionTypes,
	conditionStatus metav1.ConditionStatus,
	reason string,
	message string,
) error {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Updating %s/%s condition \"%s\"",
		res.GetObjectKind().GroupVersionKind().Kind, res.GetName(), conditionType))

	//logger.Info("Original conditions", "conditions", conditions)
	meta.SetStatusCondition(res.GetConditions(), metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		ObservedGeneration: res.GetGeneration(),
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		Message:            message,
	})
	//logger.Info("after conditions", "conditions", res.GetConditions())

	// Save the object into Kubernetes, and return any error that might have happened.
	return r.UpdateStatus(ctx, res)
}

func (b *BaseReconciler) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = log.FromContext(ctx)
	}
	return b.logger
}
