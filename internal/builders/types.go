package builders

import "errors"

// ErrTemplateDoesNotExist indicates that the TargetTemplate for the Access
// Request does not exist and therefore the Access Request cannot be satisified.
var ErrTemplateDoesNotExist = errors.New("template does not exist")

// ErrRequestDurationInvalid indicates that the requested access duration is an
// invalid time string.
var ErrRequestDurationInvalid = errors.New("access request duration invalid")

var ErrRequestDurationTooLong = errors.New(
	"access request duration longer than template maximum duration",
)

var ErrRequestExpired = errors.New("access expired")
