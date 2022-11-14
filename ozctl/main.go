package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	kubeconfig    string
	namespace     string
	configContext string
)

func main() {
	// 1. Create a flags instance.
	configFlags := genericclioptions.NewConfigFlags(true)

	// 2. Create a cobra command.
	cmd := &cobra.Command{
		Use: "ozctl",
		Run: func(cmd *cobra.Command, args []string) {

			// 4. Get client config from the flags.
			config, _ := configFlags.ToRESTConfig()

			// 5. Create a client-go instance for config.
			client, _ := kubernetes.NewForConfig(config)

			vinfo, _ := client.Discovery().ServerVersion()
			fmt.Println(vinfo)
		},
	}

	// 3. Register flags with cobra.
	configFlags.AddFlags(cmd.PersistentFlags())

	_ = cmd.Execute()
}
