package requestcontroller

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

var _ = Describe("RequestReconciler", Ordered, func() {
	/*
		fetchRequestObject() tests
	*/
	Context("fetchRequestObject()", func() {
		var (
			ctx        = context.Background()
			ns         *v1.Namespace
			request    *v1alpha1.ExecAccessRequest
			reconciler *RequestReconciler
		)

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testutil.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Should have an ExecAccessRequest built to test against")
			request = &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fetchrequestobject-test",
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
				Client:                 k8sClient,
				Scheme:                 k8sClient.Scheme(),
				APIReader:              k8sClient,
				RequestType:            &v1alpha1.ExecAccessRequest{},
				Builder:                &mockBuilder{},
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
				reconciler.RequestType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			err := reconciler.fetchRequestObject(rctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("fetchRequestObject() should fail if invalid request", func() {
			rctx := newRequestContext(
				ctx,
				reconciler.RequestType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "invalid",
						Namespace: request.GetNamespace(),
					},
				},
			)
			err := reconciler.fetchRequestObject(rctx)
			Expect(err).To(HaveOccurred())
		})
	})
})
