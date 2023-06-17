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

var _ = Describe("Validator Handler", func() {
	It("validateCreate with username matching request should succeed", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				UserInfo: authenticationv1.UserInfo{
					Username: "foo-user",
				},
				Object: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusOK)))
		Expect(resp.Allowed).Should(BeTrue())
	})
	It("validateCreate with non-matching request should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				UserInfo: authenticationv1.UserInfo{
					Username: "user",
				},
				Object: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusForbidden)))
		Expect(resp.Allowed).Should(BeFalse())
	})

	It("validateUpdate with username matching request should succeed", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				UserInfo: authenticationv1.UserInfo{
					Username: "foo-user",
				},
				Object: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
				OldObject: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusOK)))
		Expect(resp.Allowed).Should(BeTrue())
	})
	It("validateUpdate with non-matching request should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				UserInfo:  authenticationv1.UserInfo{},
				Object: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
				OldObject: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"bar-user\"}"),
				},
			},
		})
		Expect(resp.Result.Code).To(Equal(int32(http.StatusForbidden)))
		Expect(resp.Allowed).To(BeFalse())
	})
	It("validateUpdate with invalid object should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object: runtime.RawExtension{
					Raw: []byte("junk"),
				},
				OldObject: runtime.RawExtension{
					Raw: []byte("junk"),
				},
			},
		})
		Expect(resp.Result.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(resp.Allowed).To(BeFalse())
	})
	It("validateUpdate with invalid oldObject should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object: runtime.RawExtension{
					Raw: []byte("{}"),
				},
				OldObject: runtime.RawExtension{
					Raw: []byte("junk"),
				},
			},
		})
		Expect(resp.Result.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(resp.Allowed).To(BeFalse())
	})

	It("validateDelete with username should succeed", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Delete,
				OldObject: runtime.RawExtension{
					Raw: []byte("{\"requestor\": \"foo-user\"}"),
				},
			},
		})
		Expect(resp.Result.Code).Should(Equal(int32(http.StatusOK)))
		Expect(resp.Allowed).Should(BeTrue())
	})
	It("validateDelete without username should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Delete,
				OldObject: runtime.RawExtension{
					Raw: []byte("{\"requestor\":\"\"}"),
				},
			},
		})
		Expect(resp.Result.Code).To(Equal(int32(http.StatusForbidden)))
		Expect(resp.Allowed).To(BeFalse())
	})

	It("validateDelete with invalid oldObject should fail", func() {
		obj := &TestValidator{}
		decoder := admission.NewDecoder(scheme.Scheme)
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: decoder},
		}

		resp := handler.Handle(context.TODO(), admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Delete,
				OldObject: runtime.RawExtension{
					Raw: []byte("junk"),
				},
			},
		})
		Expect(resp.Result.Code).To(Equal(int32(http.StatusBadRequest)))
		Expect(resp.Allowed).To(BeFalse())
	})

	It("should fail if decode() fails", func() {
		obj := &TestValidator{}
		handler := &admission.Webhook{
			Handler: &validatorForType{object: obj, decoder: admission.NewDecoder(scheme.Scheme)},
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
			Handler: &validatorForType{object: nil},
		}
		Expect(func() {
			handler.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Delete,
				},
			})
		}).To(Panic())
	})
})

// TestDefaulter.
var _ IContextuallyValidatableObject = &TestValidator{}

type TestValidator struct {
	Requestor string `json:"requestor,omitempty"`
}

var testValidatorGVK = schema.GroupVersionKind{
	Group:   "foo.test.org",
	Version: "v1",
	Kind:    "TestValidator",
}

func (d *TestValidator) GetObjectKind() schema.ObjectKind { return d }
func (d *TestValidator) DeepCopyObject() runtime.Object {
	return &TestValidator{
		Requestor: d.Requestor,
	}
}

func (d *TestValidator) GroupVersionKind() schema.GroupVersionKind {
	return testValidatorGVK
}

func (d *TestValidator) SetGroupVersionKind(_ schema.GroupVersionKind) {}

var _ runtime.Object = &TestValidatorList{}

type TestValidatorList struct{}

func (*TestValidatorList) GetObjectKind() schema.ObjectKind { return nil }
func (*TestValidatorList) DeepCopyObject() runtime.Object   { return nil }

func (d *TestValidator) ValidateCreate(req admission.Request) (warnings admission.Warnings, err error) {
	if d.Requestor != req.UserInfo.DeepCopy().Username {
		return nil, errors.New("must have userinfo context")
	}
	return nil, nil
}

func (d *TestValidator) ValidateDelete(_ admission.Request) (warnings admission.Warnings, err error) {
	if d.Requestor == "" {
		return nil, errors.New("cannot delete")
	}
	return nil, nil
}

func (d *TestValidator) ValidateUpdate(_ admission.Request, oldObj runtime.Object) (warnings admission.Warnings, err error) {
	old := oldObj.(*TestValidator)
	if d.Requestor != old.Requestor {
		return nil, errors.New("requestor field immutable")
	}
	return nil, nil
}
