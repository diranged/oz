package cmd

import (
	"fmt"
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
)

var getCmd = &cobra.Command{
	Use:   "get <resource> ...options",
	Short: "Get an existing Access Request or Template",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the client
		builder := resource.NewBuilder(kubeConfigFlags)

		// Get the object or list of objects
		obj, err := builder.
			// Scheme teaches the Builder how to instantiate resources by names.
			WithScheme(scopedScheme, api.GroupVersion).
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

		printr := printers.NewTypeSetter(scopedScheme).
			ToPrinter(printers.NewTablePrinter(printers.PrintOptions{
				Wide:          true,
				WithNamespace: true,
				WithKind:      true,
			}))

		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			cmd.Printf(`Error: %s`, err)
			os.Exit(1)
		}
	},
}

func init() {
	kubeConfigFlags.AddFlags(getCmd.Flags())
	rootCmd.AddCommand(getCmd)
}
