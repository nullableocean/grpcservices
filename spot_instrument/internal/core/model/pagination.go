package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type PaginationData struct {
	Markets       []*Market
	HasNext       bool
	NextPageToken PageToken
}

var (
	separator = "|"
)

type PaginationCursor struct {
	MarketName string
	MarketUuid string
}

func (c PaginationCursor) Encode() PageToken {
	if c.MarketName == "" && c.MarketUuid == "" {
		return PageToken{}
	}
	data, err := json.Marshal(c)
	if err != nil {
		return PageToken{}
	}

	return PageToken{Token: base64.URLEncoding.EncodeToString(data)}
}

type PageToken struct {
	Token string
}

func (t PageToken) Decode() (PaginationCursor, error) {
	if t.Empty() {
		return PaginationCursor{}, nil
	}
	data, err := base64.URLEncoding.DecodeString(t.Token)
	if err != nil {
		return PaginationCursor{}, fmt.Errorf("invalid base64: %w", err)
	}

	var cursor PaginationCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return PaginationCursor{}, fmt.Errorf("invalid token json: %w", err)
	}

	return cursor, nil
}

func (t PageToken) Empty() bool {
	return t.Token == ""
}
