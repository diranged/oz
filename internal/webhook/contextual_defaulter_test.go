package webhook

import (
	"context"
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Defaulter Handler", func() {
	It("should return mutated object with username in create", func() {
		obj := &TestDefaulter{}
		decoder, _ := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &defaulterForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				UserInfo: authenticationv1.UserInfo{
					Username: "foo-user",
				},
				Object: runtime.RawExtension{
					Raw: []byte("{}"),
				},
			},
		})
		Expect(len(resp.Patches)).To(Equal(1))
		Expect(
			string(resp.Patch),
		).To(Equal("[{\"op\":\"add\",\"path\":\"/requestor\",\"value\":\"foo-user\"}]"))

		Expect(resp.Result.Code).Should(Equal(int32(http.StatusOK)))
		Expect(resp.Allowed).Should(BeTrue())
	})

	It("should return ok if received delete verb in defaulter handler", func() {
		obj := &TestDefaulter{}
		handler := &admission.Webhook{
			Handler: &defaulterForType{object: obj},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Delete,
				OldObject: runtime.RawExtension{
					Raw: []byte("{}"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusOK)))
		Expect(resp.Allowed).Should(BeTrue())
	})
	It("should fail if decode() fails", func() {
		obj := &TestDefaulter{}
		handler := &admission.Webhook{
			Handler: &defaulterForType{object: obj},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				OldObject: runtime.RawExtension{
					Raw: []byte("junk"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusBadRequest)))
	})

	It("should panic if no object passed in", func() {
		handler := &admission.Webhook{
			Handler: &defaulterForType{object: nil},
		}
		Expect(func() {
			handler.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Delete,
				},
			})
		}).To(Panic())
	})

	It("should fail if default() returns error", func() {
		obj := &TestDefaulter{}
		decoder, _ := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &defaulterForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				UserInfo: authenticationv1.UserInfo{
					Username: "",
				},
				Object: runtime.RawExtension{
					Raw: []byte("{}"),
				},
			},
		})
		Expect(len(resp.Patches)).To(Equal(0))
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusForbidden)))
		Expect(resp.Allowed).Should(BeFalse())
	})
})

// TestDefaulter.
var _ IContextuallyDefaultableObject = &TestDefaulter{}

type TestDefaulter struct {
	Requestor string `json:"requestor,omitempty"`
}

var testDefaulterGVK = schema.GroupVersionKind{
	Group:   "foo.test.org",
	Version: "v1",
	Kind:    "TestDefaulter",
}

func (d *TestDefaulter) GetObjectKind() schema.ObjectKind { return d }
func (d *TestDefaulter) DeepCopyObject() runtime.Object {
	return &TestDefaulter{
		Requestor: d.Requestor,
	}
}

func (d *TestDefaulter) GroupVersionKind() schema.GroupVersionKind {
	return testDefaulterGVK
}

func (d *TestDefaulter) SetGroupVersionKind(_ schema.GroupVersionKind) {}

var _ runtime.Object = &TestDefaulterList{}

type TestDefaulterList struct{}

func (*TestDefaulterList) GetObjectKind() schema.ObjectKind { return nil }
func (*TestDefaulterList) DeepCopyObject() runtime.Object   { return nil }

func (d *TestDefaulter) Default(req admission.Request) error {
	if req.UserInfo.Username != "" {
		d.Requestor = req.UserInfo.Username
		return nil
	}

	return errors.New("must have userinfo context")
}
