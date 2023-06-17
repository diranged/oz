package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// IContextuallyValidatableObject implements a similar pattern to the
// [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime/tree/v0.15.0/pkg/webhook)
// webhook pattern. The difference is that the `Default()` function is not only
// supplied the request resource, but also the request context in the form of
// an
// [`admission.Request`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/webhook/admission/webhook.go#L42C1-L65)
// object.
//
// Modified from https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/webhook/admission/defaulter_custom.go#L31-L34
type IContextuallyValidatableObject interface {
	runtime.Object
	// ValidateCreate validates the object on creation.
	// The optional warnings will be added to the response as warning messages.
	// Return an error if the object is invalid.
	ValidateCreate(req admission.Request) (warnings admission.Warnings, err error)

	// ValidateUpdate validates the object on update. The oldObj is the object before the update.
	// The optional warnings will be added to the response as warning messages.
	// Return an error if the object is invalid.
	ValidateUpdate(req admission.Request, old runtime.Object) (warnings admission.Warnings, err error)

	// ValidateDelete validates the object on deletion.
	// The optional warnings will be added to the response as warning messages.
	// Return an error if the object is invalid.
	ValidateDelete(req admission.Request) (warnings admission.Warnings, err error)
}

// RegisterContextualValidator leverages many of the patterns and code from the
// Controller-Runtime Admission package, but is one level _less_ abstracted.
// Rather than calling the `Default()` function on the target resource type,
func RegisterContextualValidator(
	obj IContextuallyValidatableObject,
	mgr ctrl.Manager,
) error {
	// Get the GroupVersionKind for the target schema object.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return err
	}
	path := generateValidatePath(gvk)

	// Create a Webhook{} resource with our Handler.
	mwh := &admission.Webhook{
		Handler: &validatorForType{object: obj, decoder: admission.NewDecoder(mgr.GetScheme())},
	}

	// Insert the path into the webhook server and point it at our mutating
	// webhook handler. This must take place before the default controller
	// NewWebhookManagedBy().Complete() function is called.
	mgr.GetWebhookServer().Register(path, mwh)

	return nil
}

// A validatorForType mimics the
// [`validatorForType`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/webhook/admission/validator_custom.go)
// code, but understands to pass the `admission.Request` object into the `Default()` function.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/webhook/admission/validator_custom.go#L57-L61
type validatorForType struct {
	object  IContextuallyValidatableObject
	decoder *admission.Decoder
}

// Handle manages the inbound request from the API server. It's responsible for
// decoding the request into an
// [`admission.Request`](https://pkg.go.dev/k8s.io/api/admission/v1#AdmissionRequest)
// object, calling the `Default()` function on that object, and then returning
// back the patched response to the API server.
// Handle handles admission requests.
//
// revive:disable:cyclomatic Replication of existing code in Controller-Runtime
func (h *validatorForType) Handle(_ context.Context, req admission.Request) admission.Response {
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/webhook/admission/validator.go#L69-L74
	if h.decoder == nil {
		panic("decoder should never be nil")
	}
	if h.object == nil {
		panic("object should never be nil")
	}

	// Get the object in the request
	obj := h.object.DeepCopyObject().(IContextuallyValidatableObject)

	var err error
	var warnings []string

	switch req.Operation {
	case v1.Connect:
		// No validation for connect requests.
		// TODO(vincepri): Should we validate CONNECT requests? In what cases?
	case v1.Create:
		if err = h.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		warnings, err = obj.ValidateCreate(req)
	case v1.Update:
		oldObj := obj.DeepCopyObject()

		err = h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		warnings, err = obj.ValidateUpdate(req, oldObj)
	case v1.Delete:
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		err = h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		warnings, err = obj.ValidateDelete(req)
	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unknown operation %q", req.Operation))
	}

	if err != nil {
		var apiStatus apierrors.APIStatus
		if errors.As(err, &apiStatus) {
			return validationResponseFromStatus(false, apiStatus.Status()).WithWarnings(warnings...)
		}
		return admission.Denied(err.Error()).WithWarnings(warnings...)
	}
	return admission.Allowed("").WithWarnings(warnings...)
}
