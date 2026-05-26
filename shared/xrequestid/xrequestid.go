package xrequestid

import "github.com/google/uuid"

func NewXRequestId() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return ""
	}

	return uuid.String()
}
