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
					Name: testutil.RandomString(8),
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
				Client:                 k8sClient,
				Scheme:                 k8sClient.Scheme(),
				APIReader:              k8sClient,
				RequestType:            &v1alpha1.ExecAccessRequest{},
				Builder:                builder,
				ReconciliationInterval: time.Minute,
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

		It("Reconcile() should work", func() {
			// Make the Mock return success on GetTemplate()
			builder.getTemplateErr = nil
			builder.getTemplateResp = &v1alpha1.ExecAccessTemplate{}

			// Make the Mock return success on GetAccessDuration()
			builder.getDurationErr = nil
			builder.getDurationResp = time.Hour

			// Make the Mock return success on CreateAccessResources
			builder.createResourcesErr = nil
			builder.createResourcesResp = "Role XYZ created"

			// Make the mock return success on AccessResourcesAreReady
			builder.accessResourcesAreReadyErr = nil
			builder.accessResourcesAreReadyResp = true

			result, err := reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			// VERIFY: The result is that we WILL requeue in a few minutes
			Expect(result.RequeueAfter).To(Equal(reconciler.ReconciliationInterval))
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
			By("Checking the resulting conditions")

			// ConditionRequestDurationsValid = True
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionRequestDurationsValid.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// ConditionTargetTemplateExists = True
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionTargetTemplateExists.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// ConditionAccessResourcesCreated = True
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionAccessResourcesCreated.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// ConditionAccessResourcesReady = True
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionAccessResourcesReady.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// Ready Status was set to true
			Expect(request.Status.IsReady()).To(BeTrue())
		})

		It("Reconcile() should not requeue if verifyDuration returns an error", func() {
			// Make the Mock return success on GetAccessDuration()
			builder.getDurationErr = fmt.Errorf("Failed: %w", builders.ErrRequestDurationInvalid)
			builder.getDurationResp = time.Duration(0)

			result, err := reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			// VERIFY: The result is that we will NOT requeue
			Expect(result.Requeue).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionRequestDurationsValid = False
			By("Checking the resulting conditions")
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionRequestDurationsValid.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusReasonBadRequest)))
			Expect(cond.Message).To(Equal("Failed: access request duration invalid"))
		})

		It("Reconcile() should requeue if verifyAccessResources returns an error", func() {
			// Make the Mock return success on GetTemplate()
			builder.getTemplateErr = nil
			builder.getTemplateResp = &v1alpha1.ExecAccessTemplate{}

			// Make the Mock return success on GetAccessDuration()
			builder.getDurationErr = nil
			builder.getDurationResp = time.Hour

			// Make the Mock return success on CreateAccessResources
			builder.createResourcesErr = nil
			builder.createResourcesResp = "Role XYZ created"

			// Make the mock return failure on AccessResourcesAreReady
			builder.accessResourcesAreReadyErr = nil
			builder.accessResourcesAreReadyResp = false

			result, err := reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			// VERIFY: The result is that we will requeue
			Expect(result.RequeueAfter).To(Equal(DefaultVerifyResourcesRequeueInterval))
			Expect(err).ToNot(HaveOccurred())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionRequestDurationsValid = True
			By("Checking the resulting conditions")
			cond := meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionRequestDurationsValid.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// ConditionAccessResourcesCreated = True
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionAccessResourcesCreated.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

			// ConditionAccessResourcesReady = False
			cond = meta.FindStatusCondition(
				*request.GetStatus().GetConditions(),
				v1alpha1.ConditionAccessResourcesReady.String(),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal("NotYetReady"))
		})

		It("Reconcile() should requeue if isAccessExpired returns an error", func() {
			// Make the Mock return success on GetTemplate()
			builder.getTemplateErr = nil
			builder.getTemplateResp = &v1alpha1.ExecAccessTemplate{}

			// Make the Mock return valid results
			builder.getDurationErr = nil
			builder.getDurationResp = time.Duration(0)

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
			Expect(err).ToNot(HaveOccurred())

			_, err = reconciler.Reconcile(
				ctx,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      request.GetName(),
						Namespace: request.GetNamespace(),
					},
				},
			)
			// VERIFY: There should be no error here
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
