package xrequestid

import "github.com/google/uuid"

const (
	XREQUEST_ID_KEY = "x-request-id"
)

func NewXRequestId() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return ""
	}

	return uuid.String()
}
