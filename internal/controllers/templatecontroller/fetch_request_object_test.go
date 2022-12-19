package templatecontroller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("TemplateReconciler", Ordered, func() {
	Context("fetchRequestObject()", func() {
		var (
			ctx        = context.Background()
			ns         *v1.Namespace
			template   *v1alpha1.ExecAccessTemplate
			reconciler *TemplateReconciler
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
					Name:      "fetchrequestobject-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "junk",
					},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			By("Creating the RequestReconciler")
			reconciler = &TemplateReconciler{
				Client:                 k8sClient,
				Scheme:                 k8sClient.Scheme(),
				TemplateType:           &v1alpha1.ExecAccessTemplate{},
				ReconciliationInterval: 0,
			}
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("fetchRequestObject() should work", func() {
			rctx := newRequestContext(
				ctx,
				reconciler.TemplateType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      template.GetName(),
						Namespace: template.GetNamespace(),
					},
				},
			)
			err := reconciler.fetchRequestObject(rctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("fetchRequestObject() should fail if invalid template", func() {
			rctx := newRequestContext(
				ctx,
				reconciler.TemplateType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "invalid",
						Namespace: template.GetNamespace(),
					},
				},
			)
			err := reconciler.fetchRequestObject(rctx)
			Expect(err).To(HaveOccurred())
		})
	})
})
