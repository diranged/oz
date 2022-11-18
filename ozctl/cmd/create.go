package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <resource> ...options",
	Short: "Command used to create an Access Request",
	Long: `This command creates the Access Request objects for you and waits until they are
	available.

	Eg:
	  $ ozctl create ExecAccessRequest --target some-template
	  ...

	  $ ozctl create AccessRequest --target some-template
	  ...
	`,
	Args: cobra.NoArgs,
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
