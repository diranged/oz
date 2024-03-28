package cmd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

var (
	// Holder of the optional --target-pod flag
	targetPod string

	// Holder for the value of the --duration flag
	duration = "1h"

	// The prefix used in the Metadata.Name field for the ExecAccessRequest object.
	requestNamePrefix = "unknown"

	// Time to wait for ExecAccessRequest to be approved and ready for use.
	waitTime = "10s"
)

var createExecAccessRequestExample = `
By default, an ExecAccessRequest will randomly select a target Pod for you:
$ ozctl create ExecAccessRequest <existing template>
...

You can optionally target a specific Pod:
$ ozctl create ExecAccessRequest <existing template> --targetPod my-existing-pod
...
`

// createAccessRequestCmd represents the create command
var createExecAccessRequestCmd = &cobra.Command{
	Aliases: []string{"execaccessrequest", "execaccessrequests", "exec-access-request", "exec"},
	Use:     "ExecAccessRequest <ExecAccessTemplate Name>",
	Short:   "Create ExecAccessRequest resources",
	Example: createExecAccessRequestExample,
	Args:    cobra.MinimumNArgs(1),

	// Static validation of the inputs - cannot be used to set state in the Run function.
	PreRunE: func(_ *cobra.Command, _ []string) error {
		// Request name prefix must start with letters a-z, can contain dashes, and must end in a
		// letter or number.
		re, err := regexp.Compile(`^[a-z][a-z0-9-][a-z0-9]+`)
		if err != nil {
			return err
		}
		if !re.MatchString(requestNamePrefix) {
			return fmt.Errorf("invalid request name prefix: %s", requestNamePrefix)
		}

		// Verify the waitTime syntax
		_, err = time.ParseDuration(waitTime)
		if err != nil {
			return fmt.Errorf("invalid time supplied: %s", waitTime)
		}

		return nil
	},

	// Do the thing
	Run: func(cmd *cobra.Command, args []string) {
		// The template must be the first argument.
		template := args[0]

		// Get our k8s client and namespace
		_, namespace := getKubeClient()

		// Create a dynamically named request template
		req := &api.ExecAccessRequest{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ExecAccessRequest",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", requestNamePrefix),
				Namespace:    namespace,
			},
			Spec: api.ExecAccessRequestSpec{
				TemplateName: template,
				Duration:     duration,
				TargetPod:    targetPod,
			},
		}

		// Verify that the target template exists proactively before creating the resource
		verifyTemplate(cmd, req)

		// Create the request resource itself now
		createAccessRequest(cmd, req)

		// Wait until the access request is ready
		waitForAccessRequest(cmd, req)
	},
}

func init() {
	createExecAccessRequestCmd.Flags().
		StringVarP(&targetPod, "target-pod", "p", "", "Optional name of a specific target pod to request access for")
	createExecAccessRequestCmd.Flags().
		StringVarP(&duration, "duration", "D", "", "Duration for the access request to be valid. Valid time units are: ns, us, ms, s, m, h.")
	createExecAccessRequestCmd.Flags().
		StringVarP(&waitTime, "wait", "w", "1m", "Duration to wait for the access request to be fully ready. Valid time units are: ns, us, ms, s, m, h.")
	createExecAccessRequestCmd.Flags().
		StringVarP(&requestNamePrefix, "request-name", "N", usernameEnv, "Prefix name to use when creating the `ExecAccessRequest` objects.")

	kubeConfigFlags.AddFlags(createExecAccessRequestCmd.Flags())

	createCmd.AddCommand(createExecAccessRequestCmd)
}
