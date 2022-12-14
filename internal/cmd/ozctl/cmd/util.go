package cmd

import (
	"github.com/fatih/color"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Define the different logging colors and names that we use. This makes it
// simple to produce output for users that is easy to read, and ensures we can
// easily change the colors as we need to in the future in one place.
//
// Example usage:
//
//	cmd.Println(logNotice("this is a faint message"))
//	cmd.Println(fmt.Sprintf(logSuccess("This message worked: %s"), someVar))
var (
	// Used for non-critical information that most of the time can be ignored.
	// Eg, "Initializing ..."
	logNotice = color.New(color.Faint).SprintFunc()

	// Success messages should be rare - they're bright and green and are
	// intended to be the most important message a user sees (that is not a
	// failure). Eg, "Your pod is available here: ..."
	logSuccess = color.New(color.FgGreen).SprintFunc()

	// Error messages are bright red and bold.
	logError = color.New(color.Bold, color.FgRed).SprintFunc()

	// Warning messages do not cause a failure - but are important to call out.
	// Soft yellow.
	logWarning = color.New(color.Faint, color.FgYellow).SprintFunc()
)

func getKubeClient() (cl client.Client, ns string) {
	kubeRestCfg, _ := kubeConfigFlags.ToRESTConfig()
	rawCl, _ := client.New(kubeRestCfg, client.Options{})
	ns = getDefaultKubeNamespace(kubeConfigFlags)
	cl = client.NewNamespacedClient(rawCl, ns)
	return cl, ns
}
