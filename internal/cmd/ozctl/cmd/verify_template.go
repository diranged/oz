package cmd

import (
	"os"

	"github.com/spf13/cobra"

	api "github.com/diranged/oz/internal/api/v1alpha1"
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

	if outputFormat == OutputFormatText {
		cmd.Printf(accessRequestInitMsg, req.GetTemplateName(), requestNamePrefix)
		cmd.Printf(verifyingTemplateExistsMsg, req.GetTemplateName(), req.GetNamespace())
	}

	// Verify the template exists
	_, err := req.GetTemplate(cmd.Context(), client)
	if err != nil {
		cmd.Printf(verifyingTemplateExistsFailedMsg, err)
		os.Exit(1)
	}
}
