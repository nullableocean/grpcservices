package errs

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrIncorrectData = errors.New("incorrect data")
	ErrNotAllowed    = errors.New("not allowed")
	ErrCantUpdate    = errors.New("cant update")
	ErrAlreadyExists = errors.New("already exists")

	ErrIdempotencyInternal    = errors.New("idempotency error")
	ErrIdempotencyProcessing  = errors.New("already processing")
	ErrIdempotencyKeyNotFound = errors.New("key not found")

	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternal            = errors.New("internal database error")
	ErrTransactionFailed   = errors.New("transaction failed")
)
