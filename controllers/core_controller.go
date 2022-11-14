package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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

func (b *BaseReconciler) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = log.FromContext(ctx)
	}
	return b.logger
}
