package service

import "errors"

var (
	ErrNotFound    = errors.New("order not found")
	ErrInvalidData = errors.New("invalid data")
)
