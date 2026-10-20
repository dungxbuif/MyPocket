package planning

import "errors"

var ErrValidation = errors.New("planning validation failed")
var ErrForbidden = errors.New("planning object forbidden")
var ErrVersionConflict = errors.New("planning version conflict")
