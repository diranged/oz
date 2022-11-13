package controllers

const (
	fieldSelectorMetadataName string = "metadata.name"
	fieldSelectorStatusPhase  string = "status.phase"
)

type RequestConditionTypes string

const (
	// Access Requests
	ConditionTargetTemplateExists RequestConditionTypes = "TargetTemplateExists"
	ConditionTargetPodExists      RequestConditionTypes = "TargetPodExists"

	// Access Templates
	ConditionTargetRefExists RequestConditionTypes = "TargetRefExists"
)
