package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidData  = errors.New("invalid data")
	ErrAccessDenied = errors.New("access denied")
	ErrAlreadyExist = errors.New("already exist")

	ErrNotAllowedMarket = fmt.Errorf("%w:market not allowed for user", ErrInvalidData)
	ErrNotAllowed       = fmt.Errorf("%w:not allowed for user", ErrAccessDenied)

	ErrStatusUnavailable = errors.New("order status unavailable")
)
