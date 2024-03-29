package controllers

const (
	// DefaultReconciliationInterval defines the number of minutes inbetween regular scheduled
	// checks of the target resources that our controllers are managing.
	DefaultReconciliationInterval int = 5

	// PodWaitReconciliationInterval is how long between attemps to check
	// whether or not a Target Pod has come up.
	PodWaitReconciliationInterval int = 5

	// EventRecorderName is the name of the Conroller used during Event recording
	EventRecorderName string = "Oz"
)
