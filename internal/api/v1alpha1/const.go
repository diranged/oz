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

const (
	// FieldSelectorMetadataName refers to the metadata.name field on an
	// object, and is used during the creation of the K8S API Client as one of
	// the fields we want to index.
	FieldSelectorMetadataName string = "metadata.name"

	// FieldSelectorStatusPhase refers to the status.phase field on an
	// object, and is used during the creation of the K8S API Client as one of
	// the fields we want to index.
	FieldSelectorStatusPhase string = "status.phase"
)
