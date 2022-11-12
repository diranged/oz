package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type CoreController interface {
	GetLogger(ctx context.Context) logr.Logger
}

type BaseController struct {
	CoreController
	logger logr.Logger
}

func (b *BaseController) GetLogger(ctx context.Context) logr.Logger {
	if (b.logger == logr.Logger{}) {
		// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
		b.logger = ctrllog.FromContext(ctx)
	}
	return b.logger
}
