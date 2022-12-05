package controllers

import (
	"context"
	"fmt"
	"strconv"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers/builders"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

// SetReconciliationInterval sets the BaseReconciler.ReconciliationInterval value to the
// DEFAULT_RECONCILIATION_INTERVAL if it was not pre-populated.
func (r *BaseReconciler) SetReconciliationInterval() {
	if r.ReconcililationInterval == 0 {
		r.ReconcililationInterval = DefaultReconciliationInterval
	}
}

// refetch uses the "consistent client" (non-caching) to retreive the latest state of the object into the
// supplied object reference. This is critical to avoid "the object has been modified; please apply
// your changes to the latest version and try again" errors when updating object status fields.
func (r *BaseReconciler) refetch(ctx context.Context, obj client.Object) (*client.Object, error) {
	if err := r.APIReader.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// UpdateStatus pushes the client.Object.Status field into Kubernetes if it has been updated, and
// then takes care of calling Refetch() to re-populate the object pointer with the updated object
// revision from Kubernetes.
//
// This wrapper makes it much easier to update the Status field of an object iteratively throughout
// a reconciliation loop.
func (r *BaseReconciler) updateStatus(ctx context.Context, res api.ICoreResource) error {
	logger := r.getLogger(ctx)

	// Update the status, handle failure.
	logger.V(2).
		Info("Pre Obj Json", "resourceVersion", res.GetResourceVersion(), "json", builders.ObjectToJSON(res))
	if err := r.Status().Update(ctx, res); err != nil {
		logger.Error(err, "Failed to update status")
		return err
	}

	// Re-fetch the object when we're done to make sure we are working with the latest version
	if _, err := r.refetch(ctx, res); err != nil {
		logger.Error(err, "Failed to refetch object")
		return err
	}

	logger.V(2).
		Info("Post Obj Json", "resourceVersion", res.GetResourceVersion(), "json", builders.ObjectToJSON(res))

	return nil
}

// updateCondition provides a simple way to update the .Status.Conditions field of a given resource. The resource
// must match the ResourceWithConditions interface - which exposes the GetConditions() method.
//
// When an updateCondition() call is made, we retrieve the current list of conditions first from the request object.
// From there, we insert in a new Condition into the resource.
// Finally we call the UpdateStatus() function to push the update to Kubernetes.
func (r *BaseReconciler) updateCondition(
	ctx context.Context,
	res api.ICoreResource,
	conditionType OzResourceConditionTypes,
	conditionStatus metav1.ConditionStatus,
	reason string,
	message string,
) error {
	logger := r.getLogger(ctx)
	logger.V(1).
		Info(fmt.Sprintf("Updating condition %s to %s", conditionType, conditionStatus))

	meta.SetStatusCondition(res.GetStatus().GetConditions(), metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		ObservedGeneration: res.GetGeneration(),
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		Message:            message,
	})

	// Save the object into Kubernetes, and return any error that might have happened.
	return r.updateStatus(ctx, res)
}

// SetReadyStatus flips the Status.Ready field to true or false. This is used at the end of a reconciliation loop
// when all of the conditions of the resource are known to have been populated. If all Conditions are in the
// ConditionSuccess status, then Status.Ready is set to true. Otherwise, it is set to False.
//
// Status.Ready is used by the 'ozctl' commandline tool to inform users when their access request
// has been approved and configured.
func (r *BaseReconciler) setReadyStatus(ctx context.Context, res api.ICoreResource) error {
	logger := r.getLogger(ctx)
	logger.V(1).Info("Checking final condition state")

	// Default to everything being ready. We'll iterate though all conditions and then flip this to false if any
	// of those conditions are not true.
	ready := true

	// Get the pointer to the conditions list
	conditions := res.GetStatus().GetConditions()

	// Iterate. If any are not true, then we flip the ready flag to false.
	for _, cond := range *conditions {
		if cond.Status != metav1.ConditionTrue {
			ready = false
		}
	}

	// Save the flag, and update the object. Return the result of the object update (if its an error).
	logger.Info(fmt.Sprintf("Setting ready state to %s", strconv.FormatBool(ready)))
	res.GetStatus().SetReady(ready)
	return r.updateStatus(ctx, res)
}

func (r *BaseReconciler) getLogger(ctx context.Context) logr.Logger {
	if (r.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		r.logger = log.FromContext(ctx)
	}
	return r.logger
}

// ignoreStatusUpdatesAndDeletion filters out reconcile requests where only the
// Status was updated, or on Deletes.
//
// **Deletes**
// On Deletes, we don't need to do any cleanup because we make sure to use
// OwnerReferences that force Kubernetes to handle the cleanup for us.
//
// **Status Updates**
// Our Reconcile() loops make many updates mid-reconcile to the status fields
// of the objects. Doing this can cause all kinds of re-runs of the reconciler
// at a high rate - mostly when they are not desired.
//
// Using this predicate filter means that the Reconcile() loops must be well
// tested and include their own automatic requeue-after settings.
//
// https://sdk.operatorframework.io/docs/building-operators/golang/references/event-filtering/
func ignoreStatusUpdatesAndDeletion() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}
