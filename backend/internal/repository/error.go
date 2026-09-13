package repository

import "errors"

var ErrNotFound = errors.New("record not found")
var ErrInvalidGoogleProfile = errors.New("invalid google profile")
