package requestcontroller

import (
	"context"
	"errors"
	"time"

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
		verifyTemplate() Tests
	*/
	Context("verifyAccessResources()", func() {
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
					Name:      "verifyduration-test",
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
				Client:                  k8sClient,
				Scheme:                  k8sClient.Scheme(),
				APIReader:               k8sClient,
				RequestType:             &v1alpha1.ExecAccessRequest{},
				Builder:                 builder,
				ReconcilliationInterval: 0,
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

		It("verifyAccessResources() should return error if resource creation fails", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.createResourcesErr = errors.New("failed")
			builder.createResourcesResp = ""

			shouldEndReconcile, _, err := reconciler.verifyAccessResources(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeTrue())

			// VERIFY: Yes, requeue with error
			Expect(err.Error()).To(Equal("failed"))

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: ConditionAccessResourcesCreated = False
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessResourcesCreated.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusFailure)))
		})

		It("verifyAccessResources() should return if access resources are not ready", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.createResourcesErr = nil
			builder.createResourcesResp = "Role-XXX created"

			// Make the Mock return false to AccessResourcesAreReady()
			builder.accessResourcesAreReadyErr = nil
			builder.accessResourcesAreReadyResp = false

			shouldEndReconcile, result, err := reconciler.verifyAccessResources(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeTrue())

			// VERIFY: Yes, result{} contains a delay
			Expect(result.RequeueAfter).To(Equal(DefaultVerifyResourcesRequeueInterval))

			// VERIFY: No, no error will trigger an immediate requeue
			Expect(err).ToNot(HaveOccurred())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: ConditionAccessResourcesCreated = True
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessResourcesCreated.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// VERIFY: ConditionAccessResourcesReady = False
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessResourcesReady.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("NotYetReady"))
		})

		It(
			"verifyAccessResources() should return if failure getting access resource state",
			func() {
				// Make the Mock return an unexpected error on getAccesssDuration()
				builder.createResourcesErr = nil
				builder.createResourcesResp = "Role-XXX created"

				// Make the Mock return false to AccessResourcesAreReady()
				builder.accessResourcesAreReadyErr = errors.New("unexpected failure")
				builder.accessResourcesAreReadyResp = false

				shouldEndReconcile, _, err := reconciler.verifyAccessResources(rctx, template)

				// VERIFY: Yes, end the reconcile
				Expect(shouldEndReconcile).To(BeTrue())

				// VERIFY: Yes, the error is returned
				Expect(err).To(HaveOccurred())

				// Refetch our Request object... reconiliation has mutated its
				// .Status fields.
				By("Refetching our Request...")
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      request.Name,
					Namespace: request.Namespace,
				}, request)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: ConditionAccessResourcesCreated = True
				cond := meta.FindStatusCondition(
					*request.GetStatus().GetConditions(),
					string(v1alpha1.ConditionAccessResourcesCreated.String()),
				)
				Expect(cond).ToNot(BeNil())
				Expect(cond.Status).To(Equal(metav1.ConditionTrue))
				Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

				// VERIFY: ConditionAccessResourcesReady = False
				cond = meta.FindStatusCondition(
					*request.GetStatus().GetConditions(),
					string(v1alpha1.ConditionAccessResourcesReady.String()),
				)
				Expect(cond).ToNot(BeNil())
				Expect(cond.Status).To(Equal(metav1.ConditionFalse))
				Expect(cond.Reason).To(Equal("NotYetReady"))
			},
		)

		It("verifyAccessResources() should return are ready", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.createResourcesErr = nil
			builder.createResourcesResp = "Role-XXX created"

			// Make the Mock return false to AccessResourcesAreReady()
			builder.accessResourcesAreReadyErr = nil
			builder.accessResourcesAreReadyResp = true

			shouldEndReconcile, result, err := reconciler.verifyAccessResources(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeFalse())

			// VERIFY: Yes, result{} contains a delay
			Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

			// VERIFY: No, no error will trigger an immediate requeue
			Expect(err).ToNot(HaveOccurred())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: ConditionAccessResourcesCreated = True
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessResourcesCreated.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// VERIFY: ConditionAccessResourcesReady = True
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessResourcesReady.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))
		})
	})
})
