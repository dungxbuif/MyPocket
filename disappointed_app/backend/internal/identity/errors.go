package identity

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrValidation = errors.New("identity validation failed")
var ErrForbidden = errors.New("forbidden")
