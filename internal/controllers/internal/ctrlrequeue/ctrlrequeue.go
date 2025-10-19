// SPDX-License-Identifier: Apache-2.0

// Package ctrlrequeue provides helper functions with clear names for informing
// the controller when to requeue (or not) reconciliations.
package ctrlrequeue

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Requeue represents that a request should be requeued for further processing.
func Requeue(requeue bool) (ctrl.Result, error) {
	if requeue {
		return ctrl.Result{RequeueAfter: time.Nanosecond}, nil
	}
	return ctrl.Result{}, nil
}

// RequeueError represents that a request should be requeued for further
// processing due to an error.
func RequeueError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueAfterError represents that a request should be requeued for further
// processing after the given interval has passed due to an error.
func RequeueAfterError(interval time.Duration, err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: interval}, err
}

// RequeueAfter represents that a request should be requeued for further
// processing after the given interval has passed.
func RequeueAfter(interval time.Duration) (ctrl.Result, error) {
	return RequeueAfterError(interval, nil)
}

// RequeueImmediately represents that a request should be requeued
// immediately for further processing.
func RequeueImmediately() (ctrl.Result, error) {
	return Requeue(true)
}

// NoRequeue represents that a request shouldn't be requeued for further processing.
func NoRequeue() (ctrl.Result, error) {
	return RequeueError(nil)
}

// ShouldRequeue returns true if we should requeue the request for reconciliation.
func ShouldRequeue(result ctrl.Result, err error) bool {
	return err != nil || result.RequeueAfter > 0
}
