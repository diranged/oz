// Package conditions may be temporary
package conditions

// OzResourceConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
// used throughout the AccessRequest and AccessTemplate reconcilers.
type OzResourceConditionTypes string

const (
	// ConditionDurationsValid is used by both AccessTemplate and AccessRequest resources. It
	// indicates whether or not the various duration fields are valid.
	ConditionDurationsValid OzResourceConditionTypes = "AccessDurationsValid"

	// ConditionTargetTemplateExists indicates that the Access Request is pointing to a valid Access
	// Template.
	ConditionTargetTemplateExists OzResourceConditionTypes = "TargetTemplateExists"

	// ConditionAccessStillValid is continaully updated based on whether or not the Access Request
	// has timed out.
	ConditionAccessStillValid OzResourceConditionTypes = "AccessStillValid"

	// ConditionAccessResourcesCreated indicates whether or not the target access request resources
	// have been properly created.
	ConditionAccessResourcesCreated OzResourceConditionTypes = "AccessResourcesCreated"

	// ConditionAccessResourcesReady indicates that all of the "access
	// resources" (eg, a Pod) are up and in the ready state.
	ConditionAccessResourcesReady OzResourceConditionTypes = "AccessResourcesReady"

	// ConditionAccessMessage is used to record
	ConditionAccessMessage OzResourceConditionTypes = "AccessMessage"

	// ConditionTargetRefExists indicates whether or not an AccessTemplate is pointing to a valid
	// Controller.
	ConditionTargetRefExists OzResourceConditionTypes = "TargetRefExists"
)
