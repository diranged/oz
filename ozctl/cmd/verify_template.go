package cmd

import (
	"os"

	api "github.com/diranged/oz/api/v1alpha1"
	"github.com/spf13/cobra"
)

var accessRequestInitMsg = logNotice(`Initiating Access Request...
  Template Name: %s
  Request Name Prefix: %s`)

var verifyingTemplateExistsMsg = logNotice(`
Verifying Template %s exists (ns: %s)...
`)

var verifyingTemplateExistsFailedMsg = logError(`
Error: - Invalid --template name flag passed in:
  %s
`)

func verifyTemplate(cmd *cobra.Command, req api.IRequestResource) {
	client, _ := getKubeClient()
	cmd.Printf(accessRequestInitMsg, req.GetTemplateName(), requestNamePrefix)

	// Verify the template exists
	cmd.Printf(verifyingTemplateExistsMsg, req.GetTemplateName(), req.GetNamespace())
	_, err := req.GetTemplate(cmd.Context(), client)
	if err != nil {
		cmd.Printf(verifyingTemplateExistsFailedMsg, err)
		os.Exit(1)
	}
}
