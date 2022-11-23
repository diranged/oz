package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

func createAccessRequest(cmd *cobra.Command, req api.IRequestResource) {
	// Get our Kubernetes Client
	client, _ := getKubeClient()

	// Pretty-print the type of object we're creating...
	reqKind := req.GetObjectKind().GroupVersionKind().GroupKind().Kind
	cmd.Printf(logNotice("Creating %s... "), reqKind)

	// Make the calls to create the request
	if err := client.Create(cmd.Context(), req); err != nil {
		fmt.Printf(
			logError("Error - Creating %s failed:\n  %s\n"),
			reqKind,
			err,
		)
	}
	cmd.Printf(logNotice("%s created!\n"), req.GetName())
}

func waitForAcessRequest(cmd *cobra.Command, req api.IRequestResource) {
	// Get our Kubernetes Client
	client, _ := getKubeClient()

	// Wait until we are either fully succesful, or we've hit our timeout.
	//
	// Newline intentionally missing.
	cmd.Printf(logNotice("Waiting for %s to be ready"), req.GetName())

	// Create a timeout context... we'll use this to bail out of our loop after waitTime has been hit.
	waitDuration, _ := time.ParseDuration(waitTime)
	waitCtx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer cancel()
	for {
		// At the beginning of each loop, update the client object from the API. If we see an
		// error, log it .. but just continue and try again.
		if err := client.Get(cmd.Context(), types.NamespacedName{
			Name:      req.GetName(),
			Namespace: req.GetNamespace(),
		}, req); err != nil {
			cmd.Printf(logWarning("\nError updating request status: %s\n"), err)
			continue
		}

		// Check the status
		if req.GetStatus().IsReady() {
			cmd.Printf(successMsg, req.GetStatus().GetAccessMessage())
			break
		}

		if waitCtx.Err() != nil {
			fmt.Printf(logError("Error - timed out waiting for %s to be ready"), req.GetName())
			for _, cond := range *req.GetStatus().GetConditions() {
				cmd.Printf(
					"Condition %s, State: %s, Reason: %s, Message: %s\n",
					cond.Type,
					cond.Status,
					cond.Reason,
					cond.Message,
				)
			}
			os.Exit(1)
		}

		// See if we've run out of time or not. If we have, bail out.
		cmd.Print(logNotice("."))
		time.Sleep(time.Duration(1 * time.Second))
	}
}
