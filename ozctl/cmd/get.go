package cmd

import (
	"fmt"
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
)

var flag string

var ourScheme *runtime.Scheme

var getCmd = &cobra.Command{
	Use:   "get <resource> ...options",
	Short: "Get an existing Access Request or Template",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Guaranteed to exist due to cobra.MinimumNArgs(1) above.
		resourceType := args[0]
		fmt.Printf("Type: %s\n", resourceType)

		// generate our scheme
		scheme, _ := api.SchemeBuilder.Build()

		// Not guaranteed.
		var resourceName string
		if len(args) >= 2 {
			resourceName = args[1]
			fmt.Printf("Name: %s\n", resourceName)
		}

		// Get the client
		builder := resource.NewBuilder(kubeConfigFlags)

		// Get the object or list of objects
		obj, err := builder.
			// Scheme teaches the Builder how to instantiate resources by names.
			WithScheme(scheme, api.GroupVersion).
			// Where to look up.
			NamespaceParam(getDefaultKubeNamespace(kubeConfigFlags)).
			// Supplied as arg 0
			ResourceTypeOrNameArgs(true, args...).
			// Do look up, please.
			Do().
			// Convert the result to a runtime.Object
			Object()
		if err != nil {
			panic(err.Error())
		}

		printr := printers.NewTypeSetter(scheme).ToPrinter(&printers.HumanReadablePrinter{})
		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	kubeConfigFlags.AddFlags(getCmd.Flags())
	rootCmd.AddCommand(getCmd)
}
