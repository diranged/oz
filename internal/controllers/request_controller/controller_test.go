package request_controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("RequestReconciler", Ordered, func() {
	/*
		Reconcile() tests
	*/
	Context("Reconcile()", func() {
		var (
			ctx        = context.Background()
			ns         *v1.Namespace
			request    *v1alpha1.ExecAccessRequest
			reconciler *RequestReconciler
			builder    = &mockBuilder{}
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

			By("Should have an ExecAccessRequest built to test against")
			request = &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reconcile-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					// Our mockBuilder will ignore this
					TemplateName: "bogus",
				},
			}
			err = k8sClient.Create(ctx, request)
			Expect(err).ToNot(HaveOccurred())

			By("Creating the RequestReconciler")
			reconciler = &RequestReconciler{
				Client:                  k8sClient,
				Scheme:                  k8sClient.Scheme(),
				APIReader:               k8sClient,
				RequestType:             &v1alpha1.ExecAccessRequest{},
				Builder:                 builder,
				ReconcilliationInterval: 0,
			}
		})

		It("Reconcile() should return if the Request object is gone", func() {
			result, err := reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "missing",
						Namespace: request.GetNamespace(),
					},
				},
			)
			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: No Requeue
			Expect(result.Requeue).To(BeFalse())
		})

		It("Should Reconcile", func() {
			// Make the Mock return success on VerifyTemplate()
			builder.getTemplateResp = nil

			_, err := reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			Expect(err).ToNot(HaveOccurred())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The condition was updated appropriately
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionTargetTemplateExists.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(metav1.StatusSuccess))
			Expect(cond.Message).To(Equal("Found Target Template"))
		})
	})
})
