package e2e

import (
	//nolint:golint
	//nolint:revive

	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/diranged/oz/controllers"
	"github.com/diranged/oz/test/utils"
	. "github.com/onsi/ginkgo/v2"

	//nolint:golint
	//nolint:revive
	. "github.com/onsi/gomega"
)

var _ = Describe("oz-controller", Ordered, func() {
	projectDir, _ := utils.GetProjectDir()

	var (
		err error

		templateSuccessConditions = []controllers.OzResourceConditionTypes{
			controllers.ConditionDurationsValid,
			controllers.ConditionTargetRefExists,
		}
		requestSuccessConditions = []controllers.OzResourceConditionTypes{
			controllers.ConditionDurationsValid,
			controllers.ConditionTargetTemplateExists,
			controllers.ConditionAccessStillValid,
			controllers.ConditionAccessResourcesCreated,
		}

		deploymentTemplate = filepath.Join(projectDir, "examples/deployment.yaml")
	)

	BeforeAll(func() {
		By("Creating target Deployment for tests")
		EventuallyWithOffset(1, func() error {
			cmd := exec.Command("kubectl", "apply", "-f", deploymentTemplate, "-n", namespace)
			_, err = utils.Run(cmd)
			return err
		}, time.Minute, time.Second).Should(Succeed())
		EventuallyWithOffset(1, func() error {
			cmd := exec.Command(
				"kubectl", "wait", "-f", deploymentTemplate, "-n", namespace, "--timeout=1s",
				"--for=condition=Available",
			)
			_, err = utils.Run(cmd)
			return err
		}, (5 * time.Minute), time.Second).Should(Succeed())
	})

	AfterAll(func() {
		By("Removing test target deployment")
		cmd := exec.Command("kubectl", "apply", "-f", deploymentTemplate, "-n", namespace)
		_, _ = utils.Run(cmd)
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
				_, err = utils.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range templateSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", template, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = utils.Run(cmd)
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
				_, err = utils.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			// Verify it is Ready
			for _, cond := range requestSuccessConditions {
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command(
						"kubectl", "wait", "-f", request, "-n", namespace, "--timeout=1s",
						fmt.Sprintf("--for=condition=%s", cond),
					)
					_, err = utils.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())
			}
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
					_, err = utils.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())

				// Verify it is Ready
				for _, cond := range templateSuccessConditions {
					EventuallyWithOffset(1, func() error {
						cmd := exec.Command(
							"kubectl", "wait", "-f", template, "-n", namespace, "--timeout=1s",
							fmt.Sprintf("--for=condition=%s", cond),
						)
						_, err = utils.Run(cmd)
						return err
					}, time.Minute, time.Second).Should(Succeed())
				}
			})

			//
			// AccessRequest tests are next - create the PodAccessRquest and wait until they have had
			// their various Conditions marked True, indicating that the Oz Controller has processed
			// them properly.
			//
			It("Should allow creating the Access Request", func() {
				By("Creating ExecAccessRequest")
				// Create the Resource
				EventuallyWithOffset(1, func() error {
					cmd := exec.Command("kubectl", "apply", "-f", request, "-n", namespace)
					_, err = utils.Run(cmd)
					return err
				}, time.Minute, time.Second).Should(Succeed())

				// Verify it is Ready
				for _, cond := range requestSuccessConditions {
					EventuallyWithOffset(1, func() error {
						cmd := exec.Command(
							"kubectl", "wait", "-f", request, "-n", namespace, "--timeout=1s",
							fmt.Sprintf("--for=condition=%s", cond),
						)
						_, err = utils.Run(cmd)
						return err
					}, time.Minute, time.Second).Should(Succeed())
				}
			})
		})
	})
})
