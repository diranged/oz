package podaccessbuilder

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("RequestReconciler", Ordered, func() {
	Context("AreAccessResourcesReady()", func() {
		var (
			ctx     = context.Background()
			ns      *corev1.Namespace
			request *v1alpha1.PodAccessRequest
			pod     *corev1.Pod
			builder = PodAccessBuilder{}
		)

		// Override the retry timeout so it won't take 30s for a failed test
		defaultReadyWaitTime = 100 * time.Millisecond
		defaultReadyWaitInterval = 10 * time.Millisecond

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: utils.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a Pod to reference for the test")
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx:latest",
						},
					},
				},
			}
			err = k8sClient.Create(ctx, pod)
			Expect(err).To(Not(HaveOccurred()))

			By("Should have an PodAccessRequest built to test against")
			request = &v1alpha1.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "createaccessresource-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.PodAccessRequestSpec{
					TemplateName: "bogus",
				},
			}
			err = k8sClient.Create(ctx, request)
			Expect(err).ToNot(HaveOccurred())

			request.Status.PodName = pod.GetName()
			err = k8sClient.Status().Update(ctx, request)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("AccessResoucesAreReady() should fail the first time", func() {
			// Execute our waiter...
			ret, err := builder.AccessResourcesAreReady(
				ctx,
				k8sClient,
				request,
				&v1alpha1.PodAccessTemplate{},
			)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: The returned ready state is True
			Expect(ret).To(BeFalse())
		})

		It("AccessResoucesAreReady() should succeed", func() {
			// Mock out the Pod status to be ready
			pod.Status.Phase = corev1.PodRunning
			err := setPodReadyCondition(
				ctx,
				pod,
				corev1.ConditionTrue,
				metav1.StatusSuccess,
				"Pod is running",
			)
			Expect(err).ToNot(HaveOccurred())

			// Execute our waiter...
			ret, err := builder.AccessResourcesAreReady(
				ctx,
				k8sClient,
				request,
				&v1alpha1.PodAccessTemplate{},
			)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: The returned ready state is True
			Expect(ret).To(BeTrue())
		})

		It("AccessResoucesAreReady() should fail if the pod has no conditions", func() {
			// Mock out the Pod status to be ready
			pod.Status.Phase = corev1.PodRunning
			pod.Status.Conditions = []corev1.PodCondition{}
			err := k8sClient.Status().Update(ctx, pod)
			Expect(err).ToNot(HaveOccurred())

			// Execute our waiter...
			ret, err := builder.AccessResourcesAreReady(
				ctx,
				k8sClient,
				request,
				&v1alpha1.PodAccessTemplate{},
			)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: The returned ready state is False
			Expect(ret).To(BeFalse())
		})

		It("AccessResoucesAreReady() should fail if the pod is never ready", func() {
			// Mock out the Pod status to be ready
			pod.Status.Phase = corev1.PodRunning
			err := setPodReadyCondition(
				ctx,
				pod,
				corev1.ConditionFalse,
				metav1.StatusFailure,
				"Pod is not yet running",
			)
			Expect(err).ToNot(HaveOccurred())

			// Execute our waiter...
			ret, err := builder.AccessResourcesAreReady(
				ctx,
				k8sClient,
				request,
				&v1alpha1.PodAccessTemplate{},
			)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: The returned ready state is False
			Expect(ret).To(BeFalse())
		})

		It("AccessResoucesAreReady() should fail immediately if the pod is missing", func() {
			// Delete the pod
			err := k8sClient.Delete(ctx, pod)
			Expect(err).ToNot(HaveOccurred())

			// Execute our waiter...
			ret, err := builder.AccessResourcesAreReady(
				ctx,
				k8sClient,
				request,
				&v1alpha1.PodAccessTemplate{},
			)

			// VERIFY: Not Found Error
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("not found"))

			// VERIFY: The returned ready state is False
			Expect(ret).To(BeFalse())
		})
		It(
			"AccessResoucesAreReady() should fail immediately if the status.PodName is not set",
			func() {
				request.Status.PodName = ""

				// Execute our waiter...
				ret, err := builder.AccessResourcesAreReady(
					ctx,
					k8sClient,
					request,
					&v1alpha1.PodAccessTemplate{},
				)

				// VERIFY: No error returned
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("status.podName not yet set"))

				// VERIFY: The returned ready state is False
				Expect(ret).To(BeFalse())
			},
		)
	})
})

func setPodReadyCondition(
	ctx context.Context,
	pod *corev1.Pod,
	status corev1.ConditionStatus,
	reason string,
	message string,
) error {
	pod.Status.Conditions = []corev1.PodCondition{
		{
			Type:   corev1.PodReady,
			Status: status,
			LastProbeTime: metav1.Time{
				Time: time.Now(),
			},
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
			Reason:  reason,
			Message: message,
		},
	}
	return k8sClient.Status().Update(ctx, pod)
}
