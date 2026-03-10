package errs

import "errors"

var (
	ErrInvalidData          = errors.New("invalid data")
	ErrAlreadyProcessed     = errors.New("order already processed")
	ErrAlreadyProcessing    = errors.New("order in processing")
	ErrDontWantProcessOrder = errors.New("dont want process order :)")
)
