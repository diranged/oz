package bldutil

import (
	"bytes"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateAccessCommand templates an access command string. The template can
// reference `{{ .Metadata }}` (the target pod's ObjectMeta) and
// `{{ .ClientKubeContext }}` (the kubeconfig context the request was created
// in, populated by `ozctl`; empty string when the request was applied as raw
// YAML).
func CreateAccessCommand(
	cmdString string,
	resource metav1.ObjectMeta,
	clientKubeContext string,
) (string, error) {
	type data struct {
		Metadata          metav1.ObjectMeta
		ClientKubeContext string
	}
	d := data{
		Metadata:          resource,
		ClientKubeContext: clientKubeContext,
	}

	tmpl, err := template.New("accessCommand").Parse(cmdString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}
