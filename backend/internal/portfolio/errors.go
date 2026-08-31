package portfolio

import "errors"

var ErrValidation = errors.New("portfolio validation failed")
var ErrForbidden = errors.New("portfolio object forbidden")
var ErrConflict = errors.New("portfolio version conflict")
var ErrOversell = errors.New("portfolio trade oversells position")
