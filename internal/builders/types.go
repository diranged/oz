package builders

import "errors"

// ErrTemplateDoesNotExist indicates that the TargetTemplate for the Access
// Request does not exist and therefore the Access Request cannot be satisified.
var ErrTemplateDoesNotExist = errors.New("template does not exist")
