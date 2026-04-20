package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const fakeKubeconfig = `apiVersion: v1
kind: Config
current-context: kube-current
clusters:
- name: c1
  cluster:
    server: https://example.invalid
contexts:
- name: kube-current
  context:
    cluster: c1
    user: u1
- name: kube-other
  context:
    cluster: c1
    user: u1
users:
- name: u1
  user: {}
`

func writeTempKubeconfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(fakeKubeconfig), 0o600); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	return path
}

func TestKubeContextFromFlags(t *testing.T) {
	kubeconfigPath := writeTempKubeconfig(t)
	empty := ""
	override := "kube-other"

	tests := []struct {
		name  string
		setup func() *genericclioptions.ConfigFlags
		want  string
	}{
		{
			name: "context flag wins over kubeconfig current-context",
			setup: func() *genericclioptions.ConfigFlags {
				f := genericclioptions.NewConfigFlags(false)
				f.KubeConfig = &kubeconfigPath
				f.Context = &override
				return f
			},
			want: "kube-other",
		},
		{
			name: "falls back to kubeconfig current-context when flag is empty",
			setup: func() *genericclioptions.ConfigFlags {
				f := genericclioptions.NewConfigFlags(false)
				f.KubeConfig = &kubeconfigPath
				f.Context = &empty
				return f
			},
			want: "kube-current",
		},
		{
			name: "returns empty string when no kubeconfig is loadable",
			setup: func() *genericclioptions.ConfigFlags {
				// Point at a non-existent kubeconfig. We also clear KUBECONFIG
				// from the env so the loader can't fall back to it.
				t.Setenv("KUBECONFIG", "/nonexistent/path/that/does/not/exist")
				bogus := "/nonexistent/path/that/does/not/exist"
				f := genericclioptions.NewConfigFlags(false)
				f.KubeConfig = &bogus
				f.Context = &empty
				return f
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kubeContextFromFlags(tt.setup())
			if got != tt.want {
				t.Errorf("kubeContextFromFlags() = %q, want %q", got, tt.want)
			}
		})
	}
}
