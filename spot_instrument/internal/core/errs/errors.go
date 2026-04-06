package errs

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAllowed    = errors.New("not found")
	ErrIncorrectData = errors.New("incorrect data")
)
