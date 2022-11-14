package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type OzReconciler interface {
	GetLogger(ctx context.Context) logr.Logger
	Refetch(ctx context.Context, obj client.Object) *client.Object
}

type BaseReconciler struct {
	OzReconciler
	client.Client

	Scheme *runtime.Scheme
	logger logr.Logger
}

func (b *BaseReconciler) Refetch(ctx context.Context, obj client.Object) error {
	err := b.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, obj)
	return err
}

func (b *BaseReconciler) UpdateStatus(ctx context.Context, obj client.Object) error {
	logger := b.GetLogger(ctx)

	// Update the status, handle failure.
	if err := b.Status().Update(ctx, obj); err != nil {
		logger.Error(err, fmt.Sprintf("Failed to update %s status", obj.GetObjectKind()))
		return err
	}

	// Refetch the object
	b.Refetch(ctx, obj)

	return nil
}

func (b *BaseReconciler) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = ctrllog.FromContext(ctx)
	}
	return b.logger
}
