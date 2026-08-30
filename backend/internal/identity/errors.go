package identity

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrForbidden = errors.New("forbidden")
