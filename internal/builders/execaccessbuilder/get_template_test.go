package execaccessbuilder

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("ExecAccessBuilder", Ordered, func() {
	Context("GetTemplate()", func() {
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

		It("GetTemplate() should work", func() {
			request := &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "verifytemplate-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: template.GetName(),
				},
			}
			builder := ExecAccessBuilder{}
			tmpl, err := builder.GetTemplate(ctx, k8sClient, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(tmpl.GetName()).To(Equal(template.GetName()))
		})

		It("GetTemplate() should throw TemplateDoesNotExist", func() {
			request := &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "verifytemplate-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: "missing",
				},
			}
			builder := ExecAccessBuilder{}
			_, err := builder.GetTemplate(ctx, k8sClient, request)
			Expect(err).To(Equal(builders.ErrTemplateDoesNotExist))
		})

		It("GetTemplate() should throw unexpected errors", func() {
			request := &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "verifytemplate-missing",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{},
			}
			builder := ExecAccessBuilder{}
			_, err := builder.GetTemplate(ctx, k8sClient, request)
			Expect(err.Error()).To(Equal("resource name may not be empty"))
		})
	})
})
