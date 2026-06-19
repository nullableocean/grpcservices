package xrequestid

import "github.com/google/uuid"

func NewXRequestId() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}
