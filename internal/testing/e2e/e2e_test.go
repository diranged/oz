package e2e

import (
	//nolint:golint
	//nolint:revive

	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	//nolint:golint
	//nolint:revive
	. "github.com/onsi/gomega"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("oz-controller", Ordered, func() {
	projectDir, _ := testutil.GetProjectDir()

	var (
		err error

		templateSuccessConditions = []v1alpha1.TemplateConditionTypes{
			v1alpha1.ConditionTemplateDurationsValid,
			v1alpha1.ConditionTargetRefExists,
		}
		execRequestSuccessConditions = []v1alpha1.RequestConditionTypes{
			v1alpha1.ConditionRequestDurationsValid,
			v1alpha1.ConditionTargetTemplateExists,
			v1alpha1.ConditionAccessStillValid,
			v1alpha1.ConditionAccessResourcesCreated,
		}
		podRequestSuccessConditions = []v1alpha1.RequestConditionTypes{
			v1alpha1.ConditionRequestDurationsValid,
			v1alpha1.ConditionTargetTemplateExists,
			v1alpha1.ConditionAccessStillValid,
			v1alpha1.ConditionAccessResourcesCreated,
			v1alpha1.ConditionAccessResourcesReady,
		}

		deploymentTemplate = filepath.Join(projectDir, "examples/deployment.yaml")
	)

	BeforeAll(func() {
		By("Creating target Deployment for tests")
		EventuallyWithOffset(1, func() error {
			cmd := exec.Command("kubectl", "apply", "-f", deploymentTemplate, "-n", namespace)
			_, err = testutil.Run(cmd)
			return err
		}, time.Minute, time.Second).Should(Succeed())
		EventuallyWithOffset(1, func() error {
			cmd := exec.Command(
				"kubectl", "wait", "-f", deploymentTemplate, "-n", namespace, "--timeout=1s",
				"--for=condition=Available",
			)
			_, err = testutil.Run(cmd)
			return err
		}, (5 * time.Minute), time.Second).Should(Succeed())
	})

	AfterAll(func() {
		By("Removing test target deployment")
		cmd := exec.Command("kubectl", "apply", "-f", deploymentTemplate, "-n", namespace)
		_, _ = testutil.Run(cmd)
	})

	Context("ExecAccessTemplate / ExecAccessRequest", func() {
		template := filepath.Join(projectDir, "examples/exec_access_template.yaml")
		request := filepath.Join(projectDir, "examples/exec_access_request.yaml")

		//
		// Initial tests - create the ExecAccessTemplate and PodAccessTemplates.
		// Wait until they have had their various Conditions marked True, indicating that the Oz
		// Controller has processed them properly.
		//
		It("Should allow creating the Access Templates", func() {
			By("Creating ExecAccessTemplate")
			// Create the Resource
			EventuallyWithOffset(1, func() error {
				cmd := exec.Command("kubectl", "apply", "-f", template, "-n", namespace)
				_, err = testutil.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range templateSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", template, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = testutil.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())
			}
		})

		//
		// AccessRequest tests are next - create the ExecAccessRquest and wait until they have had
		// their various Conditions marked True, indicating that the Oz Controller has processed
		// them properly.
		//
		It("Should allow creating the Access Request", func() {
			By("Creating ExecAccessRequest")
			// Create the Resource
			EventuallyWithOffset(1, func() error {
				cmd := exec.Command("kubectl", "apply", "-f", request, "-n", namespace)
				_, err = testutil.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range execRequestSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", request, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = testutil.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())
			}
		})
	})

	Context("PodAccessTemplate / PodAccessRequest", func() {
		template := filepath.Join(projectDir, "examples/pod_access_template.yaml")
		request := filepath.Join(projectDir, "examples/pod_access_request.yaml")

		//
		// Initial tests - create the PodAccessTemplate and PodAccessTemplates.
		// Wait until they have had their various Conditions marked True, indicating that the Oz
		// Controller has processed them properly.
		//
		It("Should allow creating the Access Templates", func() {
			By("Creating PodAccessTemplate")
			// Create the Resource
			EventuallyWithOffset(1, func() error {
				cmd := exec.Command("kubectl", "apply", "-f", template, "-n", namespace)
				_, err = testutil.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range templateSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", template, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = testutil.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())
			}
		})

		//
		// PodAccessRequest tests are next - create the PodAccessRquest and
		// wait until they have had their various Conditions marked True,
		// indicating that the Oz Controller has processed them properly.
		//
		It("Should allow creating the Access Request", func() {
			By("Creating PodAccessRequest")
			// Create the Resource
			EventuallyWithOffset(1, func() error {
				cmd := exec.Command("kubectl", "apply", "-f", request, "-n", namespace)
				_, err = testutil.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range podRequestSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", request, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = testutil.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())
			}

			By("Checking AccessMessage is valid and not empty")
			cmd := exec.Command(
				"kubectl",
				"get",
				"-f",
				request,
				"-n",
				namespace,
				"-o=jsonpath={.status.accessMessage}",
			)
			message, err := testutil.Run(cmd)
			Expect(err).To(Not(HaveOccurred()))
			Expect(
				message,
			).To(MatchRegexp("kubectl exec -ti -n oz-system deployment-example-.* -- /bin/bash"))
		})

		It("Should allow CONNECTing to the Pod", func() {
			By("Execing into the pod")
			var podName string

			EventuallyWithOffset(1, func() error {
				cmd := exec.Command(
					"kubectl", "get", "-f", request, "-n", namespace, "-o", "jsonpath='{.status.podName}'",
				)

				// Get the podname from the template. Note, it comes back wrapped in single quotes.
				podName, err = testutil.Run(cmd)
				Expect(err).To(Not(HaveOccurred()))
				Expect(podName).NotTo(BeEmpty())

				// Strip the single quotes from the podName string.
				podName = strings.Replace(podName, "'", "", -1)
				Expect(podName).NotTo(BeEmpty())

				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify that the CONNECT and validating webhook handler work
			EventuallyWithOffset(1, func() error {
				cmd := exec.Command(
					"kubectl", "exec", "-t", "-n", namespace, podName, "--", "whoami",
				)
				whoami, err := testutil.Run(cmd)
				Expect(err).To(Not(HaveOccurred()))
				Expect(
					whoami,
				).To(MatchRegexp("root"))
				return err
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
