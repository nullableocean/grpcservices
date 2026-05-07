package model

import (
	"encoding/base64"
	"fmt"
	"strings"
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

	return PageToken{
		Token: base64.URLEncoding.EncodeToString([]byte(c.MarketName + separator + c.MarketUuid)),
	}
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
		return PaginationCursor{}, fmt.Errorf("invalid page token: %w", err)
	}

	parts := strings.SplitN(string(data), separator, 2)
	if len(parts) != 2 {
		return PaginationCursor{}, fmt.Errorf("invalid page token format")
	}

	return PaginationCursor{
		MarketName: parts[0],
		MarketUuid: parts[1],
	}, nil
}

func (t PageToken) Empty() bool {
	return t.Token == ""
}
