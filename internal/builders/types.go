package builders

import "errors"

// ErrTemplateDoesNotExist indicates that the TargetTemplate for the Access
// Request does not exist and therefore the Access Request cannot be satisified.
var ErrTemplateDoesNotExist = errors.New("template does not exist")

// ErrRequestDurationInvalid indicates that the requested access duration is an
// invalid time string.
var ErrRequestDurationInvalid = errors.New("access request duration invalid")

// ErrRequestDurationTooLong indicates that the Access Request's "duration"
// field is longer than the target templates "maxDuration" field.
var ErrRequestDurationTooLong = errors.New(
	"access request duration longer than template maximum duration",
)

// ErrRequestExpired indicates that the Access Request has expired
var ErrRequestExpired = errors.New("access expired")
