package ports

import "errors"

var (
	ErrFailedClientRequest = errors.New("failed client request")
	ErrNotFound            = errors.New("not found")
	ErrService             = errors.New("service error")
	ErrAlreadyExist        = errors.New("already exist")
)
