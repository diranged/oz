package podwatcher

import (
	"github.com/diranged/oz/internal/controllers"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// example code: https://github.com/kubernetes-sigs/controller-runtime/blob/master/examples/builtins/validatingwebhook.go

// PodWatcher is a ValidatingWebhookEndpoint that receives calls from the
// Kubernetes API just before Pod's "exec" subresource is written into the
// cluster. The intention for this resource is to perform audit-logging type
// actions in the short term, and in the long term provide a more granular
// layer of security for Pod Exec access.
type PodWatcher struct {
	Client   client.Client
	decoder  admission.Decoder
	recorder record.EventRecorder
}

// NewPodWatcherRegistration creates a PodWatcher{} object and registers it at the supplied path.
func NewPodWatcherRegistration(
	mgr manager.Manager,
	path string,
) {
	hookServer := mgr.GetWebhookServer()

	hookServer.Register(
		path,
		&webhook.Admission{
			Handler: &PodWatcher{
				Client:   mgr.GetClient(),
				decoder:  *admission.NewDecoder(mgr.GetScheme()),
				recorder: mgr.GetEventRecorderFor(controllers.EventRecorderName),
			},
		},
	)
}
