package e2e

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/diranged/oz/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Make sure this matches the namespace in "config/default/kustomization.yaml"
const namespace = "oz-system"

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting Oz Operator suite\n")
	RunSpecs(t, "Oz e2e suite")
}

// Before we start the suite, pre-build the docker image, create the test namespace and get
// everything spun up.
var _ = BeforeSuite(func() {

	cmd := exec.Command("kubectl", "create", "ns", namespace)

	cmd = exec.Command("make", "docker-build")
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	cmd = exec.Command("make", "docker-load")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	cmd = exec.Command("make", "deploy")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var controllerPodName string
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
		cmd = exec.Command("kubectl", "get", "pods",
			controllerPodName,
			"-o", "jsonpath={.status.phase}",
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
})

// After the suite, undeploy the resources to clean up as much as possible.
var _ = AfterSuite(func() {
	By("tearing down the test resources")
	cmd := exec.Command("make", "undeploy")
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
})
