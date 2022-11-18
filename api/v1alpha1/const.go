package v1alpha1

// ControllerKind is a string that represents an Apps/V1 known controller kind that this codebase
// supports. This is used to limit the inputs on the AccessTemplate and ExecAccessTemplate CRDs.
type ControllerKind string

const (
	// DeploymentController maps to APIVersion: apps/v1, Kind: Deployment
	DeploymentController ControllerKind = "Deployment"

	// DaemonSetController maps to APIVersion: apps/v1, Kind: DaemonSet
	DaemonSetController ControllerKind = "DaemonSet"

	// StatefulSetController maps to APIVersion: apps/v1, Kind: StatfulSet
	StatefulSetController ControllerKind = "StatefulSet"
)
