package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// admissionRequestFor builds a minimal admission.Request carrying an operation
// and an authenticated username.
func admissionRequestFor(op admissionv1.Operation, username string) admission.Request {
	return admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: op,
			UserInfo:  authenticationv1.UserInfo{Username: username},
		},
	}
}

var _ = Describe("RequestedBy annotation", func() {
	Context("SetRequestedBy", func() {
		It("stamps the authenticated user on create", func() {
			request := &PodAccessRequest{}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Create, "dennis"))
			Expect(GetRequestedBy(request)).To(Equal("dennis"))
		})

		It("preserves any annotations that were already present", func() {
			request := &PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"unrelated": "value"},
				},
			}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Create, "dennis"))
			Expect(GetRequestedBy(request)).To(Equal("dennis"))
			Expect(request.GetAnnotations()).To(HaveKeyWithValue("unrelated", "value"))
		})

		It("overwrites a client-supplied value on create, so it cannot be spoofed", func() {
			request := &PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{AnnotationRequestedBy: "somebody-else"},
				},
			}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Create, "dennis"))
			Expect(GetRequestedBy(request)).To(Equal("dennis"))
		})

		It("removes a client-supplied value when the API server gave no identity", func() {
			request := &PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{AnnotationRequestedBy: "somebody-else"},
				},
			}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Create, ""))
			Expect(GetRequestedBy(request)).To(BeEmpty())
		})

		// The reconciler issues Updates to set owner references. If those
		// rewrote the annotation, every request would end up attributed to the
		// Oz controller's own service account.
		It("leaves the annotation alone on update", func() {
			request := &PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{AnnotationRequestedBy: "dennis"},
				},
			}
			SetRequestedBy(request, admissionRequestFor(
				admissionv1.Update, "system:serviceaccount:oz-system:oz-controller-manager",
			))
			Expect(GetRequestedBy(request)).To(Equal("dennis"))
		})

		It("does not create the annotation on update", func() {
			request := &PodAccessRequest{}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Update, "dennis"))
			Expect(GetRequestedBy(request)).To(BeEmpty())
		})

		It("works for ExecAccessRequests too", func() {
			request := &ExecAccessRequest{}
			SetRequestedBy(request, admissionRequestFor(admissionv1.Create, "dennis"))
			Expect(GetRequestedBy(request)).To(Equal("dennis"))
		})
	})

	Context("ValidateRequestedByUnchanged", func() {
		withUser := func(username string) *PodAccessRequest {
			return &PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{AnnotationRequestedBy: username},
				},
			}
		}

		It("permits an update that does not touch the annotation", func() {
			Expect(ValidateRequestedByUnchanged(withUser("dennis"), withUser("dennis"))).
				To(Succeed())
		})

		It("permits an update on an object that never had the annotation", func() {
			Expect(ValidateRequestedByUnchanged(&PodAccessRequest{}, &PodAccessRequest{})).
				To(Succeed())
		})

		It("rejects an attempt to reassign the requester", func() {
			err := ValidateRequestedByUnchanged(withUser("dennis"), withUser("somebody-else"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(AnnotationRequestedBy))
		})

		It("rejects an attempt to remove the annotation", func() {
			err := ValidateRequestedByUnchanged(withUser("dennis"), &PodAccessRequest{})
			Expect(err).To(HaveOccurred())
		})

		It("rejects an attempt to add the annotation after the fact", func() {
			err := ValidateRequestedByUnchanged(&PodAccessRequest{}, withUser("dennis"))
			Expect(err).To(HaveOccurred())
		})
	})
})
