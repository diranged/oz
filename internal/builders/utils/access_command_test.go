package bldutil

import (
	"testing"

	"github.com/diranged/oz/internal/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateAccessCommand(t *testing.T) {
	type args struct {
		cmdString         string
		resource          *v1alpha1.ExecAccessTemplate
		clientKubeContext string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				cmdString: "kubectl exec -ti -n {{ .Metadata.Namespace }} {{ .Metadata.Name }} -- /bin/sh",
				resource: &v1alpha1.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podName",
						Namespace: "namespace",
					},
				},
			},
			want:    "kubectl exec -ti -n namespace podName -- /bin/sh",
			wantErr: false,
		},
		{
			name: "client kube context is interpolated",
			args: args{
				cmdString: "kubectl --context {{ .ClientKubeContext }} exec -ti -n {{ .Metadata.Namespace }} {{ .Metadata.Name }} -- /bin/sh",
				resource: &v1alpha1.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podName",
						Namespace: "namespace",
					},
				},
				clientKubeContext: "us1",
			},
			want:    "kubectl --context us1 exec -ti -n namespace podName -- /bin/sh",
			wantErr: false,
		},
		{
			name: "empty client kube context renders as empty string",
			args: args{
				cmdString: "ctx=[{{ .ClientKubeContext }}]",
				resource: &v1alpha1.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "n"},
				},
			},
			want:    "ctx=[]",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAccessCommand(tt.args.cmdString, tt.args.resource.ObjectMeta, tt.args.clientKubeContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccessCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateAccessCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
