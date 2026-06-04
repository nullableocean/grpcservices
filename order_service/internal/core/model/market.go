package model

import "time"

type Market struct {
	UUID      string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
