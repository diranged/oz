package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type OzReconciler interface {
	GetLogger(ctx context.Context) logr.Logger
}

type BaseReconciler struct {
	OzReconciler
	client.Client

	Scheme *runtime.Scheme
	logger logr.Logger
}

func (b *BaseReconciler) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = ctrllog.FromContext(ctx)
	}
	return b.logger
}
