package service

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidData  = errors.New("invalid data")
	ErrAccessDenied = errors.New("access denied")
)
