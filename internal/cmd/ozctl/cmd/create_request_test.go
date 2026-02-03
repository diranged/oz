package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

func TestPrintOutput(t *testing.T) {
	// Create a test PodAccessRequest
	req := &api.PodAccessRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodAccessRequest",
			APIVersion: api.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-request",
			Namespace: "test-namespace",
		},
		Spec: api.PodAccessRequestSpec{
			TemplateName: "test-template",
			Duration:     "1h",
		},
		Status: api.PodAccessRequestStatus{
			CoreStatus: api.CoreStatus{
				Ready:         true,
				AccessMessage: "kubectl exec -ti test-pod -- /bin/sh",
			},
			PodName: "test-pod",
		},
	}

	tests := []struct {
		name         string
		format       string
		wantContains []string
		wantJSON     bool
		wantYAML     bool
	}{
		{
			name:   "json output",
			format: OutputFormatJSON,
			wantContains: []string{
				`"kind": "PodAccessRequest"`,
				`"name": "test-request"`,
				`"namespace": "test-namespace"`,
				`"templateName": "test-template"`,
				`"podName": "test-pod"`,
			},
			wantJSON: true,
		},
		{
			name:   "yaml output",
			format: OutputFormatYAML,
			wantContains: []string{
				"kind: PodAccessRequest",
				"name: test-request",
				"namespace: test-namespace",
				"templateName: test-template",
				"podName: test-pod",
			},
			wantYAML: true,
		},
		{
			name:   "text output",
			format: OutputFormatText,
			wantContains: []string{
				"Success",
				"kubectl exec -ti test-pod -- /bin/sh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the output format
			outputFormat = tt.format

			// Capture the output
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)

			// Call printOutput
			printOutput(cmd, req)

			output := buf.String()

			// Check that expected content is present
			for _, want := range tt.wantContains {
				if !bytes.Contains([]byte(output), []byte(want)) {
					t.Errorf("output does not contain %q\nGot:\n%s", want, output)
				}
			}

			// Verify JSON is valid
			if tt.wantJSON {
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
					t.Errorf("output is not valid JSON: %v\nGot:\n%s", err, output)
				}
			}

			// Verify YAML is valid
			if tt.wantYAML {
				var yamlData map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &yamlData); err != nil {
					t.Errorf("output is not valid YAML: %v\nGot:\n%s", err, output)
				}
			}
		})
	}
}

func TestPrintOutputDefaultIsText(t *testing.T) {
	// Reset outputFormat to default
	outputFormat = OutputFormatText

	req := &api.PodAccessRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodAccessRequest",
			APIVersion: api.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-test",
			Namespace: "default-namespace",
		},
		Spec: api.PodAccessRequestSpec{
			TemplateName: "test-template",
		},
		Status: api.PodAccessRequestStatus{
			CoreStatus: api.CoreStatus{
				Ready:         true,
				AccessMessage: "kubectl exec -ti default-pod -- /bin/sh",
			},
		},
	}

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	printOutput(cmd, req)

	output := buf.String()

	// Should contain success message (text output)
	if !bytes.Contains([]byte(output), []byte("Success")) {
		t.Errorf("default output should be text format with Success message\nGot:\n%s", output)
	}

	// Should contain the access message
	if !bytes.Contains([]byte(output), []byte("kubectl exec -ti default-pod -- /bin/sh")) {
		t.Errorf("text output missing access message\nGot:\n%s", output)
	}
}

func TestPrintOutputExecAccessRequest(t *testing.T) {
	// Test that ExecAccessRequest also works (since it implements IRequestResource)
	outputFormat = OutputFormatJSON

	req := &api.ExecAccessRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ExecAccessRequest",
			APIVersion: api.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "exec-test",
			Namespace: "exec-namespace",
		},
		Spec: api.ExecAccessRequestSpec{
			TemplateName: "exec-template",
			Duration:     "30m",
		},
		Status: api.ExecAccessRequestStatus{
			CoreStatus: api.CoreStatus{
				Ready:         true,
				AccessMessage: "kubectl exec -ti exec-pod -- /bin/bash",
			},
			PodName: "exec-pod",
		},
	}

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	printOutput(cmd, req)

	output := buf.String()

	// Should be valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		t.Errorf("ExecAccessRequest output is not valid JSON: %v\nGot:\n%s", err, output)
	}

	// Should contain ExecAccessRequest specific fields
	if !bytes.Contains([]byte(output), []byte(`"kind": "ExecAccessRequest"`)) {
		t.Errorf("JSON output missing kind field\nGot:\n%s", output)
	}
}
