package cmd

import (
	"fmt"
	"regexp"
	"time"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createPodAccessRequestExample = `
A PodAccessRequest always generates a new Pod for you to do your work in. You simply run:

$ ozctl create PodAccessRequest <existing template>
...
Success, your access request is ready! Here are your access instructions:

kubectl exec -ti -n default user-vd9r9-a217f263 -- /bin/sh
`

// createPodAccessRequestCmd represents the create command
var createPodAccessRequestCmd = &cobra.Command{
	Aliases: []string{"podaccessrequest", "podaccessrequests", "pod-access-request", "pod"},
	Use:     "PodAccessRequest <PodAccessTemplate Name>",
	Short:   "Create PodAccessRequest resources",
	Example: createPodAccessRequestExample,
	Args:    cobra.MinimumNArgs(1),

	// Static validation of the inputs - cannot be used to set state in the Run function.
	PreRunE: func(cmd *cobra.Command, args []string) error {
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
		templateName := args[0]

		// Get our k8s client and namespace
		_, namespace := getKubeClient()

		// Create a dynamically named request template
		req := &api.PodAccessRequest{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodAccessRequest",
				APIVersion: api.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", requestNamePrefix),
				Namespace:    namespace,
			},
			Spec: api.PodAccessRequestSpec{
				TemplateName: templateName,
				Duration:     duration,
			},
		}

		// Verify that the target template exists proactively before creating the resource
		verifyTemplate(cmd, req)

		// Create the request resource itself now
		createAccessRequest(cmd, req)

		// Wait until the access request is ready
		waitForAcessRequest(cmd, req)
	},
}

func init() {
	createPodAccessRequestCmd.Flags().
		StringVarP(&duration, "duration", "D", "", "Duration for the access request to be valid. Valid time units are: ns, us, ms, s, m, h.")
	createPodAccessRequestCmd.Flags().
		StringVarP(&requestNamePrefix, "request-name", "N", usernameEnv, "Prefix name to use when creating the `AccessRequest` objects.")

	kubeConfigFlags.AddFlags(createPodAccessRequestCmd.Flags())

	createCmd.AddCommand(createPodAccessRequestCmd)
}
