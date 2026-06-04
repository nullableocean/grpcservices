package errs

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAllowed    = errors.New("not allowed")
	ErrIncorrectData = errors.New("incorrect data")
)
