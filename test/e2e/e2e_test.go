package e2e

import (
	//nolint:golint
	//nolint:revive
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/diranged/oz/test/utils"
	. "github.com/onsi/ginkgo/v2"

	//nolint:golint
	//nolint:revive
	. "github.com/onsi/gomega"
)

// Make sure this matches the namespace in "config/default/kustomization.yaml"
const namespace = "oz-system"

var _ = Describe("oz", Ordered, func() {
	// BeforeAll(func() {
	// 	// Create the target namespace for the installation
	// 	By("creating manager namespace")
	// 	cmd := exec.Command("kubectl", "create", "ns", namespace)
	// 	utils.Run(cmd)
	// })

	// AfterAll(func() {
	// 	By("removing manager namespace")
	// 	cmd := exec.Command("kubectl", "delete", "ns", "--force", namespace)
	// 	_, _ = utils.Run(cmd)
	// })

	Context("Oz Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error
			projectDir, _ := utils.GetProjectDir()

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the new manager(Operator) image into Kind")
			cmd = exec.Command("make", "docker-load")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing app")
			cmd = exec.Command("make", "deploy")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name
				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}{{ if not .metadata.deletionTimestamp }}{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)
				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

			By("creating a Deployment to test access against")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "apply", "-f", filepath.Join(projectDir,
					"examples/deployment.yaml"), "-n", namespace)
				_, err = utils.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())
		})

	})

})
