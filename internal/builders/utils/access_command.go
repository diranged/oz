package bldutil

import (
	"bytes"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateAccessCommand templates an access command string,
// evaluates data from a pod.ObjectMeta
func CreateAccessCommand(cmdString string, resource metav1.ObjectMeta) (string, error) {
	type md struct {
		Metadata metav1.ObjectMeta
	}
	m := md{resource}

	tmpl, err := template.New("accessCommand").Parse(cmdString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m); err != nil {
		return "", err
	}
	return buf.String(), nil
}
