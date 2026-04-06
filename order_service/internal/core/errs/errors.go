package errs

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrIncorrectData         = errors.New("incorrect data")
	ErrNotAllowed            = errors.New("not allowed")
	ErrCantUpdate            = errors.New("cant update")
	ErrIdempotencyInternal   = errors.New("idempotency error")
	ErrIdempotencyProcessing = errors.New("already processing")
)
