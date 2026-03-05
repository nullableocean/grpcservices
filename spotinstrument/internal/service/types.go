package service

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidUUID  = errors.New("invalid market uuid")
	ErrAlreadyExist = errors.New("already exist")
)
