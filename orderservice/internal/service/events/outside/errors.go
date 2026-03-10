package outside

import "errors"

var (
	ErrEventAlreadyHandled = errors.New("event already handled")
)
