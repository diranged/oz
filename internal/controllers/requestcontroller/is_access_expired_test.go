package requestcontroller

import (
	"context"
	"time"

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
		verifyTemplate() Tests
	*/
	Context("isAccessExpireed()", func() {
		var (
			ctx        = context.Background()
			ns         *v1.Namespace
			request    *v1alpha1.ExecAccessRequest
			template   *v1alpha1.ExecAccessTemplate
			reconciler *RequestReconciler
			builder    = &mockBuilder{}
			rctx       *RequestContext
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

			By("Should have an ExecAccessTemplate to test against")
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
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
						Name:       "fake",
					},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			By("Should have an ExecAccessRequest built to test against")
			request = &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "isaccessexpired-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: template.GetName(),
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
				Builder:                builder,
				ReconciliationInterval: 0,
			}

			By("Creating the RequestContext")
			rctx = newRequestContext(
				ctx,
				reconciler.RequestType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)

			By("Populuating the rctx.obj object...")
			err = reconciler.fetchRequestObject(rctx)
			Expect(err).To(BeNil())
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("isAccessExpired() should return if no condition found", func() {
			// Overwrite the conditions in the object - so there are none to start with.
			request.Status.Conditions = []metav1.Condition{}
			err := k8sClient.Status().Update(ctx, request)
			rctx.obj = request
			Expect(err).ToNot(HaveOccurred())

			// Execute
			shouldEndReconcile, _, err := reconciler.isAccessExpired(rctx)

			// VERIFY: No, do not end
			Expect(shouldEndReconcile).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It(
			"isAccessExpired() should return if expired found, and trigger end of reconcile",
			func() {
				// Overwrite the conditions in the object - so there are none to start with.
				request.Status.Conditions = []metav1.Condition{
					{
						Type:               string(v1alpha1.ConditionAccessStillValid),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 1,
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
						Reason:  "AccessExpired",
						Message: "It is over",
					},
				}
				err := k8sClient.Status().Update(ctx, request)
				rctx.obj = request
				Expect(err).ToNot(HaveOccurred())

				// Execute
				shouldEndReconcile, _, err := reconciler.isAccessExpired(rctx)

				// VERIFY: Yes, end the reconcile
				Expect(shouldEndReconcile).To(BeTrue())
				// VERIFY: No, an error did not occur while deleting the object
				Expect(err).ToNot(HaveOccurred())

				// VERIFY: The object was deleted
				dErr := k8sClient.Get(ctx, types.NamespacedName{
					Name:      request.GetName(),
					Namespace: request.GetNamespace(),
				}, &v1alpha1.ExecAccessRequest{})
				Expect(dErr).To(HaveOccurred())
				Expect(dErr.Error()).To(MatchRegexp("not found"))
			},
		)
	})
})
