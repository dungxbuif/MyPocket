package sync

import "errors"

var ErrValidation = errors.New("sync validation failed")
var ErrConflict = errors.New("sync conflict")
