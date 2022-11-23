package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var createExample = `
# Create an ExecAccessRequest with ExecAccessTemplate "some-template"
ozctl create ExecAccessRequest --target some-template

# Create a PodAccessRequest with PodAccessTemplate "some-template"
ozctl create PodAccessRequest --target some-template
`

var createCmd = &cobra.Command{
	Use:     "create <resource> ...options",
	Short:   "Command used to create an Access Request",
	Long:    `This command creates the Access Request objects for you and waits until they are available.`,
	Example: createExample,
	Args:    cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
