// Package cmd provides all of the individual command line flags that can be passed into the 'ozctl' tool.
package cmd

import (
	"fmt"
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Variables filled out during the initial `Execute()` step. These are used throughout the various
// commands create within this CLI tool.
var (
	// A populated controller-runtime.client.Client object that can be used to make REST calls to Kubernetes.
	KubeClient client.Client

	// Shortcut to the supplied "--namespace" flag value (if supplied), or the default value from
	// the .kubeconfig. Finally defaults to "default" if nothing else is supplied.
	KubeNamespace = "default"

	//
	Username = os.Getenv("USER")
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ozctl",
	Short: "Manage Oz Access Requests and Approvals",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	verifyUsernameSet()

	// Set up common CLI flags, and then build the Kubernetes Client Object from them
	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(rootCmd.PersistentFlags())

	// Build the Kubernetes client
	kubeRestCfg, _ := configFlags.ToRESTConfig()
	KubeClient, _ = client.New(kubeRestCfg, client.Options{})

	// Make sure to add in our api schemes with the custom resources
	api.AddToScheme(scheme.Scheme)

	// Determine the default namespace
	KubeNamespace = getDefaultKubeNamespace(configFlags)

	// Set up the root command and make sure that doesn't fail.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getDefaultKubeNamespace(cf *genericclioptions.ConfigFlags) string {
	clientConfig := cf.ToRawKubeConfigLoader()
	discoveredNamespace, _, err := clientConfig.Namespace()
	if err != nil {
		return "default"
	}
	return discoveredNamespace
}

func verifyUsernameSet() error {
	if Username == "" {
		fmt.Println("ERROR: This CLI tool requires that the $USER environment be set to something")
		os.Exit(1)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {}
