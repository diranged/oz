package controllers

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

const (
	// DefaultReconciliationInterval defines the number of minutes inbetween regular scheduled
	// checks of the target resources that our controllers are managing.
	DefaultReconciliationInterval int = 5

	// PodWaitReconciliationInterval is how long between attemps to check
	// whether or not a Target Pod has come up.
	PodWaitReconciliationInterval int = 5
)
