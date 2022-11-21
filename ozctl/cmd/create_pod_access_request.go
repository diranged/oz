package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// createPodAccessRequestCmd represents the create command
var createPodAccessRequestCmd = &cobra.Command{
	Aliases: []string{"podaccessrequest", "podaccessrequests", "pod-access-request", "pod"},
	Use:     "PodAccessRequest --template <PodAccessTemplate Name>",
	Short:   "Create PodAccessRequest resources",
	Long: `This command creates PodAccessRequest resources. Example:

	By default, an PodAccessRequest will randomly select a target Pod for you:
	$ ozctl create PodAccessRequest --template <existing template>
	...

    You can optionally target a specific Pod:
	$ ozctl create PodAccessRequest --template <existing template> --targetPod my-existing-pod
	...
	`,
	Args: cobra.OnlyValidArgs,

	// Static validation of the inputs - cannot be used to set state in the Run function.
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Request name prefix must start with letters a-z, can contain dashes, and must end in a
		// letter or number.
		cmd.Print("Validating --request-name prefix... ")
		re, err := regexp.Compile(`^[a-z][a-z0-9-][a-z0-9]+`)
		if err != nil {
			return err
		}
		if !re.MatchString(requestNamePrefix) {
			return fmt.Errorf("invalid request name prefix: %s", requestNamePrefix)
		}
		cmd.Printf("valid!\n")

		// Verify the waitTime syntax
		cmd.Print("Validating --wait-time... ")
		_, err = time.ParseDuration(waitTime)
		if err != nil {
			return fmt.Errorf("invalid time supplied: %s", waitTime)
		}
		cmd.Printf("valid!\n")

		return nil
	},

	// Do the thing
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Initiating Access Request...")
		cmd.Printf("  Template Name: %s\n", template)
		cmd.Printf("  Request Name Prefix: %s\n", requestNamePrefix)
		cmd.Printf("\n")

		// Verify the template exists
		cmd.Printf("Verifying Template %s exists... ", template)
		_, err := api.GetPodAccessTemplate(cmd.Context(), KubeClient, template, KubeNamespace)
		if err != nil {
			fmt.Printf("Error - Invalid --template name flag passed in:\n  %s\n", err)
			os.Exit(1)
		}
		cmd.Printf("it does!\n")

		// Create a dynamically named request template
		req := &api.PodAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", requestNamePrefix),
				Namespace:    KubeNamespace,
			},
			Spec: api.PodAccessRequestSpec{
				TemplateName: template,
				Duration:     duration,
			},
		}

		// Create the request object
		cmd.Printf("Creating %s... ", api.PodAccessRequest{}.Kind)
		if err = KubeClient.Create(cmd.Context(), req); err != nil {
			fmt.Printf("Error - Creating %s failed:\n  %s\n", api.PodAccessRequest{}.Kind, err)
		}
		cmd.Printf("%s created!\n", req.Name)

		// Wait until we are either fully succesful, or we've hit our timeout.
		//
		// Newline intentionally missing.
		cmd.Print("Waiting for PodAccessRequest to be ready.")

		// Create a timeout context... we'll use this to bail out of our loop after waitTime has been hit.
		waitDuration, _ := time.ParseDuration(waitTime)
		waitCtx, cancel := context.WithTimeout(context.Background(), waitDuration)
		defer cancel()
		for {
			// At the beginning of each loop, update the client object from the API. If we see an
			// error, log it .. but just continue and try again.
			if err := KubeClient.Get(cmd.Context(), types.NamespacedName{
				Name:      req.GetName(),
				Namespace: req.GetNamespace(),
			}, req); err != nil {
				cmd.Printf("\nError updating request status: %s\n", err)
				continue
			}

			// Check the status
			if req.GetStatus().IsReady() {
				cmd.Println("\nSuccess, your access request is ready!")
				cmd.Printf("\nAccess your pod with: kubectl exec -ti -n %s %s\n", req.GetNamespace(), req.GetPodName())
				break
			}

			if waitCtx.Err() != nil {
				fmt.Println("\nError - timed out waiting for PodAccessRequest to be ready")
				for _, cond := range *req.GetStatus().GetConditions() {
					cmd.Printf("Condition %s, State: %s, Reason: %s, Message: %s\n", cond.Type, cond.Status, cond.Reason, cond.Message)
				}
				os.Exit(1)
			}

			// See if we've run out of time or not. If we have, bail out.
			cmd.Print(".")
			time.Sleep(time.Duration(1 * time.Second))
		}

	},
}

func init() {
	createPodAccessRequestCmd.Flags().StringVarP(&template, "template", "t", "", "Name of the AccessTemplate to request access from")
	createPodAccessRequestCmd.MarkFlagRequired("template")
	createPodAccessRequestCmd.Flags().StringVarP(&duration, "duration", "D", "", "Duration for the access request to be valid. Valid time units are: ns, us, ms, s, m, h.")
	createPodAccessRequestCmd.Flags().StringVarP(&requestNamePrefix, "request-name", "N", Username, "Prefix name to use when creating the `AccessRequest` objects.")

	createCmd.AddCommand(createPodAccessRequestCmd)
}
