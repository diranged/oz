package v1alpha1

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// AnnotationRequestedBy records the Kubernetes username that created an access
// request, as reported by the API server on the admission request.
//
// The identity is only ever known inside an admission webhook - by the time a
// request reaches the reconciler, the "who" has been lost. Stamping it onto the
// object during mutation is what makes per-user reporting (and after-the-fact
// auditing with plain kubectl) possible.
//
// Do not read this field as a security control on its own; it is set from
// authoritative API server data on create and is immutable thereafter, but it
// says who *asked* for access, not who was *granted* it. Access itself is
// granted to the template's allowedGroups.
const AnnotationRequestedBy = "crds.wizardofoz.co/requested-by"

// SetRequestedBy stamps the AnnotationRequestedBy annotation onto an object
// from the authenticated user identity on an admission request. It is intended
// to be called from a mutating (defaulting) webhook.
//
// This only acts on CREATE operations, for two reasons: the annotation should
// record the original requester rather than whoever most recently touched the
// object, and the Oz reconciler itself issues Updates against these objects
// (to set owner references), which would otherwise rewrite the annotation to
// the controller's own service account.
//
// On create the annotation is set unconditionally, overwriting any value the
// client supplied, so that a user cannot attribute their request to somebody
// else. If the API server supplied no identity at all the annotation is
// removed rather than left at a client-supplied value.
func SetRequestedBy(obj metav1.Object, req admission.Request) {
	if req.Operation != admissionv1.Create {
		return
	}

	annotations := obj.GetAnnotations()

	if req.UserInfo.Username == "" {
		delete(annotations, AnnotationRequestedBy)
		obj.SetAnnotations(annotations)
		return
	}

	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[AnnotationRequestedBy] = req.UserInfo.Username
	obj.SetAnnotations(annotations)
}

// GetRequestedBy returns the username recorded by SetRequestedBy, or an empty
// string if the object predates this feature or was created without an
// identity.
func GetRequestedBy(obj metav1.Object) string {
	return obj.GetAnnotations()[AnnotationRequestedBy]
}

// ValidateRequestedByUnchanged returns an error if an update would modify the
// AnnotationRequestedBy annotation.
//
// SetRequestedBy deliberately ignores updates so that the controller's own
// writes do not clobber the requester. That leaves the annotation editable by
// anyone with update access to the object, which would make it worthless for
// reporting - so updates to it are rejected here instead.
func ValidateRequestedByUnchanged(old, updated metav1.Object) error {
	oldValue := GetRequestedBy(old)
	newValue := GetRequestedBy(updated)
	if oldValue == newValue {
		return nil
	}
	return fmt.Errorf(
		"error - %s is an immutable annotation (%q), cannot update to %q",
		AnnotationRequestedBy,
		oldValue,
		newValue,
	)
}
