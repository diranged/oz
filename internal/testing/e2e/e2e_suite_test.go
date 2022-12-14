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
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting Oz Operator suite\n")
	RunSpecs(t, "Oz e2e suite")
}

// Before we start the suite, pre-build the docker image, create the test namespace and get
// everything spun up.
var _ = BeforeSuite(func() {
	_ = exec.Command("kubectl", "create", "ns", namespace)

	cmdRelease := exec.Command("make", "release")
	_, err := utils.Run(cmdRelease)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	cmdDockerLoad := exec.Command("make", "docker-load")
	_, err = utils.Run(cmdDockerLoad)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	cmdDeploy := exec.Command("make", "deploy")
	_, err = utils.Run(cmdDeploy)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	EventuallyWithOffset(1, func() error {
		cmd := exec.Command(
			"kubectl", "wait", "deployment",
			"-l", "control-plane=controller-manager",
			"-n", namespace, "--timeout=1s",
			"--for=condition=Available",
		)
		_, err = utils.Run(cmd)
		return err
	}, (5 * time.Minute), time.Second).Should(Succeed())
})

// After the suite, undeploy the resources to clean up as much as possible.
var _ = AfterSuite(func() {
	By("tearing down the test resources")
	cmd := exec.Command("make", "undeploy")
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
})
