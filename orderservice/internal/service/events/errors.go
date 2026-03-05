package events

import "errors"

var (
	ErrEventAlreadyHandled = errors.New("event already handled")
)
