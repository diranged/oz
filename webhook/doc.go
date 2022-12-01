// Package webhook provides a version of the controller-runtime
// [webhook](https://github.com/kubernetes-sigs/controller-runtime/tree/master/pkg/webhook)
// package. This version passes the
// [`admission.Request`](https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/webhook/admission/webhook.go#L48-L50)
// object into the `Default()`, `ValidateCreate()`, `ValidateUpdate()` and
// `ValidateDelete()` functions to provide more context to these functions for
// making their decisions.
package webhook
