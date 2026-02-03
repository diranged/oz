package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

// getOutputFormat holds the output format for the get command (table, json, yaml)
var getOutputFormat = OutputFormatTable

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

		switch getOutputFormat {
		case OutputFormatJSON:
			data, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				fmt.Printf(logError("Error marshalling to JSON: %s\n"), err)
				os.Exit(1)
			}
			cmd.Println(string(data))
		case OutputFormatYAML:
			data, err := yaml.Marshal(obj)
			if err != nil {
				fmt.Printf(logError("Error marshalling to YAML: %s\n"), err)
				os.Exit(1)
			}
			cmd.Print(string(data))
		default: // OutputFormatTable
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
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&getOutputFormat, "output", "o", OutputFormatTable,
		"Output format: table, json, or yaml")
	kubeConfigFlags.AddFlags(getCmd.Flags())
	rootCmd.AddCommand(getCmd)
}
