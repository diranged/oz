package v1alpha1

import (
	"time"
)

// AccessConfig provides a common interface for our Template structs (which implement
// ITemplateResource) for defining which entities are being granted access to a resource, and for
// how long they are granted that access.
type AccessConfig struct {
	// AllowedGroups lists out the groups (in string name form) that will be allowed to Exec into
	// the target pod.
	//
	// +kubebuilder:validation:Required
	AllowedGroups []string `json:"allowedGroups"`

	// DefaultDuration sets the default time that an access request resource will live. Must
	// be set below MaxDuration.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	//
	// +kubebuilder:default:="1h"
	DefaultDuration string `json:"defaultDuration"`

	// MaxDuration sets the maximum duration that an access request resource can request to
	// stick around.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	//
	// +kubebuilder:default:="24h"
	MaxDuration string `json:"maxDuration"`

	// AccessCommand is used to describe to the user how they can make use of their temporary access.
	// The AccessCommand can reference data from a Pod ObjectMeta.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="kubectl exec -it -n {{ .Metadata.Namespace }} {{ .Metadata.Name }} -- /bin/sh"
	AccessCommand string `json:"accessCommand"`
}

// GetAllowedGroups returns the Spec.AllowedGroups for this particular template
func (a *AccessConfig) GetAllowedGroups() []string {
	return a.AllowedGroups
}

// GetDefaultDuration parses the Spec.defaultDuration field into a time.Duration struct.
//
// Returns:
//
//	time.Duration: Populated struct (or nil, if error)
//	error: If any error occurs in the parsing, the error is returned
func (a *AccessConfig) GetDefaultDuration() (time.Duration, error) {
	return time.ParseDuration(a.DefaultDuration)
}

// GetMaxDuration parses the Spec.maxDuration field into a time.Duration struct.
//
// Returns:
//
//	time.Duration: Populated struct (or nil, if error)
//	error: If any error occurs in the parsing, the error is returned
func (a *AccessConfig) GetMaxDuration() (time.Duration, error) {
	return time.ParseDuration(a.MaxDuration)
}
