// Package cmd provides all of the individual command line flags that can be passed into the 'ozctl' tool.
package cmd

import (
	"fmt"
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
)

// Variables filled out during the initial `Execute()` step. These are used throughout the various
// commands create within this CLI tool.
var (
	Username        = os.Getenv("USER")
	kubeConfigFlags = genericclioptions.NewConfigFlags(true)
)

var rootCmd = &cobra.Command{
	Use:   "ozctl",
	Short: "Manage Oz Access Requests and Approvals",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	verifyUsernameSet()

	// Make sure to add in our api schemes with the custom resources
	api.AddToScheme(scheme.Scheme)

	// Set up the root command and make sure that doesn't fail.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func verifyUsernameSet() error {
	if Username == "" {
		fmt.Println("ERROR: This CLI tool requires that the $USER environment be set to something")
		os.Exit(1)
	}
	return nil
}

func getDefaultKubeNamespace(cf *genericclioptions.ConfigFlags) string {
	if *cf.Namespace != "" {
		return *cf.Namespace
	}

	clientConfig := cf.ToRawKubeConfigLoader()
	discoveredNamespace, _, err := clientConfig.Namespace()
	if err != nil {
		return "default"
	}
	return discoveredNamespace
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {}
