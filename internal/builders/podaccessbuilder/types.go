// Package podaccessbuilder implements the IBuilder interface for PodAccessRequest resources
package podaccessbuilder

import (
	"time"

	"github.com/diranged/oz/internal/builders"
)

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests/finalizers,verbs=update

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete;bind;escalate
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argoproj.io,resources=rollouts,verbs=get;list;watch

// defaultReadyWaitTime is the default time in which we wait for resources to
// become Ready in the AccessResourcesAreReady() method.
var defaultReadyWaitTime = 30 * time.Second

// defaultReadyWaitInterval is the time inbetween checks on the Pod status.
var defaultReadyWaitInterval = time.Second

// PodAccessBuilder implements the IBuilder interface for PodAccessRequest resources
type PodAccessBuilder struct{}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ builders.IBuilder = &PodAccessBuilder{}
	_ builders.IBuilder = (*PodAccessBuilder)(nil)
)
