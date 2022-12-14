package webhook

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Webhook", Ordered, func() {
	Context("Utils", func() {
		It("generateMutatePath()/generateValidatePath()", func() {
			gvk := schema.GroupVersionKind{
				Group:   "testGroup.io",
				Version: "v1alpha1",
				Kind:    "testKind",
			}
			Expect(generateMutatePath(gvk)).To(Equal("/mutate-testGroup-io-v1alpha1-testkind"))
			Expect(generateValidatePath(gvk)).To(Equal("/validate-testGroup-io-v1alpha1-testkind"))
		})

		It("validationResponseFromStatus()", func() {
			ret := validationResponseFromStatus(true, metav1.Status{
				Status:  "Status",
				Message: "Message",
				Reason:  "Reason",
				Code:    200,
			})
			Expect(ret.Result.Status).To(Equal("Status"))
			Expect(ret.Result.Code).To(Equal(int32(200)))
			Expect(ret.Allowed).To(BeTrue())
		})
	})
})
