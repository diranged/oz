package requestcontroller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("RequestReconciler", Ordered, func() {
	/*
		verifyTemplate() Tests
	*/
	Context("verifyDuration()", func() {
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

		It("verifyDuration() should return RequestDurationTooLong", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.getDurationErr = fmt.Errorf("Failed: %w", builders.ErrRequestDurationTooLong)
			builder.getDurationResp = time.Duration(0)

			shouldEndReconcile, result, err := reconciler.verifyDuration(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeTrue())

			// VERIFY: No, do not requeue
			Expect(result.Requeue).To(BeFalse())
			Expect(err).To(BeNil())

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
				string(v1alpha1.ConditionRequestDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("BadRequest"))

			// VERIFY: The ConditionAccessStillValid was not set either way
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessStillValid.String()),
			)
			Expect(cond).To(BeNil())
		})

		It("verifyDuration() should return RequestDurationInvalid", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.getDurationErr = fmt.Errorf("Failed: %w", builders.ErrRequestDurationInvalid)
			builder.getDurationResp = time.Duration(0)

			shouldEndReconcile, result, err := reconciler.verifyDuration(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeTrue())

			// VERIFY: No, do not requeue
			Expect(result.Requeue).To(BeFalse())
			Expect(err).To(BeNil())

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
				string(v1alpha1.ConditionRequestDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("BadRequest"))

			// VERIFY: The ConditionAccessStillValid was not set either way
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessStillValid.String()),
			)
			Expect(cond).To(BeNil())
		})

		It("verifyDuration() should return Unexpected Error and Requeue", func() {
			// Make the Mock return an unexpected error on getAccesssDuration()
			builder.getDurationErr = fmt.Errorf("Failed: Unexpected")
			builder.getDurationResp = time.Duration(0)

			shouldEndReconcile, _, err := reconciler.verifyDuration(rctx, template)

			// VERIFY: Yes, end the reconcile
			Expect(shouldEndReconcile).To(BeTrue())

			// VERIFY: Yes, requeue!
			Expect(err).To(Equal(builder.getDurationErr))

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
				string(v1alpha1.ConditionRequestDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("BadRequest"))

			// VERIFY: The ConditionAccessStillValid was not set either way
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessStillValid.String()),
			)
			Expect(cond).To(BeNil())
		})

		It("verifyDuration() should succeed, and determine the access is expired", func() {
			// Make the Mock return a duration that is definitely expired
			builder.getDurationErr = nil
			builder.getDurationResp = time.Duration(-1)

			shouldEndReconcile, _, err := reconciler.verifyDuration(rctx, template)

			// VERIFY: No, do not end the reconcile
			Expect(shouldEndReconcile).To(BeFalse())
			Expect(err).To(BeNil())

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
				string(v1alpha1.ConditionRequestDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("Success"))

			// VERIFY: The ConditionAccessStillValid was not set either way
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessStillValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("Timeout"))
		})

		It("verifyDuration() should succeed, and determine the access is still valid", func() {
			// Make the Mock return a duration that is definitely expired
			builder.getDurationErr = nil
			builder.getDurationResp = time.Hour

			shouldEndReconcile, _, err := reconciler.verifyDuration(rctx, template)

			// VERIFY: No, do not end the reconcile
			Expect(shouldEndReconcile).To(BeFalse())
			Expect(err).To(BeNil())

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
				string(v1alpha1.ConditionRequestDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("Success"))

			// VERIFY: The ConditionAccessStillValid was not set either way
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				string(v1alpha1.ConditionAccessStillValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal("Success"))
		})
	})
})
