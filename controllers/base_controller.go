package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/diranged/oz/interfaces"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OzReconciler extends the default reconciler behaviors (client.Client+Scheme) and provide some helper
// functions for refetching objects directly from the API, pushing status updates, etc.
type OzReconciler struct {
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

	// ReconciliationInterval is the time to wait inbetween re-reconciling ExecAccessRequests. This primarily matters
	// for setting the maximum time after an AccessRequest has expired that it will be purged by the controller.
	ReconcililationInterval int
}

// Set's the default reconciliation interval property - used typically at the end of the reconciliation loop.
func (b *OzReconciler) SetReconciliationInterval() {
	if b.ReconcililationInterval == 0 {
		b.ReconcililationInterval = DEFAULT_RECONCILIATION_INTERVAL
	}
}

// Refetch uses the "consistent client" (non-caching) to retreive the latest state of the object into the
// supplied object reference. This is critical to avoid "the object has been modified; please apply
// your changes to the latest version and try again" errors when updating object status fields.
func (b *OzReconciler) Refetch(ctx context.Context, obj client.Object) error {
	return b.ApiReader.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, obj)
}

func (b *OzReconciler) UpdateStatus(ctx context.Context, obj client.Object) error {
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
func (r *OzReconciler) UpdateCondition(
	ctx context.Context,
	res interfaces.OzResource,
	conditionType OzResourceConditionTypes,
	conditionStatus metav1.ConditionStatus,
	reason string,
	message string,
) error {
	logger := r.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Updating condition \"%s\"", conditionType))

	meta.SetStatusCondition(res.GetConditions(), metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		ObservedGeneration: res.GetGeneration(),
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		Message:            message,
	})

	// Save the object into Kubernetes, and return any error that might have happened.
	return r.UpdateStatus(ctx, res)
}

func (r *OzReconciler) SetReadyStatus(ctx context.Context, res interfaces.OzResource) error {
	logger := r.GetLogger(ctx)
	logger.Info("Checking final condition state")

	// Default to everything being ready. We'll iterate though all conditions and then flip this to false if any
	// of those conditions are not true.
	ready := true

	// Get the pointer to the conditions list
	conditions := res.GetConditions()

	// Iterate. If any are not true, then we flip the ready flag to false.
	for _, cond := range *conditions {
		if cond.Status != metav1.ConditionTrue {
			ready = false
		}
	}

	// Save the flag, and update the object. Return the result of the object update (if its an error).
	logger.Info(fmt.Sprintf("Setting ready state to %s", strconv.FormatBool(ready)))
	res.SetReady(ready)
	return r.UpdateStatus(ctx, res)
}

func (b *OzReconciler) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = log.FromContext(ctx)
	}
	return b.logger
}
