package execaccessbuilder

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("ExecAccessBuilder", Ordered, func() {
	Context("SetOwnerReference()", func() {
		var (
			ctx      = context.Background()
			ns       *v1.Namespace
			template *v1alpha1.ExecAccessTemplate
		)

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: utils.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Should have an ExecAccessTemplate built to test against")
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "verifytemplate-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "foo",
					},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("SetOwnerReference() should work", func() {
			By("Creating an ExecAccessRequest object")
			request := &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: template.GetName(),
				},
			}
			err := k8sClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			By("Calling the SetOwnerReference() function")
			builder := ExecAccessBuilder{}
			err = builder.SetOwnerReference(ctx, k8sClient, request, template)
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: The owner reference got set?
			Expect(len(request.ObjectMeta.OwnerReferences)).To(Equal(1))
		})

		It("SetOwnerReference() should fail if the template is invalid", func() {
			By("Creating an ExecAccessRequest object")
			invalidtemplate := &v1alpha1.ExecAccessTemplate{}
			request := &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: "missing",
				},
			}
			err := k8sClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			By("Calling the SetOwnerReference() function")
			builder := ExecAccessBuilder{}
			err = builder.SetOwnerReference(ctx, k8sClient, request, invalidtemplate)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("uid must not be empty"))
		})
	})
})
