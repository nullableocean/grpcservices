package xrequestid

import "github.com/google/uuid"

func NewXRequestId() string {
	return uuid.NewString()
}
