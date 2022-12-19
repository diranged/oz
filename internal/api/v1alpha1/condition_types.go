package v1alpha1

// IConditionType provides an interface for accepting any condition string that
// has a String() function. This simplifies the
// controllers/internal/status/update_status.go code to have a single
// UpdateStatus() function.
//
// +kubebuilder:object:generate=false
type IConditionType interface {
	String() string
}

// RequestConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
// used throughout the AccessRequest and AccessTemplate reconcilers.
type RequestConditionTypes string

const (
	// ConditionRequestDurationsValid is used by both AccessTemplate and
	// AccessRequest resources. It indicates whether or not the various
	// duration fields are valid.
	ConditionRequestDurationsValid RequestConditionTypes = "AccessDurationsValid"

	// ConditionTargetTemplateExists indicates that the Access Request is
	// pointing to a valid Access Template.
	ConditionTargetTemplateExists RequestConditionTypes = "TargetTemplateExists"

	// ConditionAccessStillValid is continaully updated based on whether or not
	// the Access Request has timed out.
	ConditionAccessStillValid RequestConditionTypes = "AccessStillValid"

	// ConditionAccessResourcesCreated indicates whether or not the target
	// access request resources have been properly created.
	ConditionAccessResourcesCreated RequestConditionTypes = "AccessResourcesCreated"

	// ConditionAccessResourcesReady indicates that all of the "access
	// resources" (eg, a Pod) are up and in the ready state.
	ConditionAccessResourcesReady RequestConditionTypes = "AccessResourcesReady"

	// ConditionAccessMessage is used to record
	ConditionAccessMessage RequestConditionTypes = "AccessMessage"
)

// String implements the fmt.Stringer interface.
func (x RequestConditionTypes) String() string { return string(x) }

// TemplateConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
// used throughout the AccessTemplate reconcilers and written to the ITemplateResource resources.
type TemplateConditionTypes string

const (
	// ConditionTemplateDurationsValid is used by both AccessTemplate and
	// AccessRequest resources. It indicates whether or not the various
	// duration fields are valid.
	ConditionTemplateDurationsValid TemplateConditionTypes = "TemplateDurationsValid"

	// ConditionTargetRefExists indicates whether or not an AccessTemplate is
	// pointing to a valid Controller.
	ConditionTargetRefExists TemplateConditionTypes = "TargetRefExists"
)

// String implements the fmt.Stringer interface.
func (x TemplateConditionTypes) String() string { return string(x) }
