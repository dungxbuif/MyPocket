package finance

import "errors"

var ErrValidation = errors.New("finance validation failed")
var ErrSystemCategoryLocked = errors.New("system category is locked")
var ErrForbidden = errors.New("finance object forbidden")
var ErrConflict = errors.New("finance version conflict")
