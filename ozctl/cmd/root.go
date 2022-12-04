// Package cmd provides all of the individual command line flags that can be passed into the 'ozctl' tool.
package cmd

import (
	"errors"
	"fmt"
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
)

// Variables filled out during the initial `Execute()` step. These are used throughout the various
// commands create within this CLI tool.
var (
	// usernamEnv stores the value of the `USER` environment variable. This is
	// used for naming the Access Requests in somewhat friendly ways - but has
	// nothing to do with whether or not access is granted via RBAC.

	usernameEnv = os.Getenv("USER")

	// kubeConfigFlags are generated once and stored here for reference by the sub commands.
	kubeConfigFlags = genericclioptions.NewConfigFlags(true)

	// scopedScheme provides a runtime.Scheme that is scoped specifically to
	// our api/v1alpha1 API only, so it does not include any other native
	// Kubernetes resource types.
	scopedScheme *runtime.Scheme
)

var rootCmd = &cobra.Command{
	Use:   "ozctl",
	Short: "Manages Oz Access Requests and Approvals",
	Long: `
Manages Oz Access Requests and Approvals.

This tool provides access to create (and approve, in the future) Access Requests
for resources within a Kubernetes cluster running the Oz RBAC Controller.
Access Requests are short-lived temporary permissions requests to manage
existing resources, or requests for dedicated short term resources (like a
temporary development Pod).

`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// https://github.com/ivanpirog/coloredcobra
	cc.Init(&cc.Config{
		RootCmd:  rootCmd,
		Headings: cc.HiCyan + cc.Bold + cc.Underline,
		Commands: cc.HiYellow + cc.Bold,
		Example:  cc.Italic,
		ExecName: cc.Bold,
		Flags:    cc.Bold,
	})

	// Sanity check
	if err := verifyUsernameSet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Populate our scopedScheme
	scopedScheme, _ = api.SchemeBuilder.Build()

	// Make sure to add in our api schemes with the custom resources
	if err := api.AddToScheme(scheme.Scheme); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set up the root command and make sure that doesn't fail.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func verifyUsernameSet() error {
	if usernameEnv == "" {
		return errors.New("this CLI tool requires that the $USER environment be set to something")
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
