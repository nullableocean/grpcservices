package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
)

type OrderList struct {
	List           []*Order
	HasNextPage    bool
	NextPageCursor *PaginationCursor
}

type OrderListFilter struct {
	Statuses    []OrderStatus
	MarketUUIDs []string
	Type        *OrderType
	CreatedFrom *time.Time
	CreatedTo   *time.Time

	PageCursor *PaginationCursor
	Limit      int
}

func (f OrderListFilter) Validate() error {
	if f.Limit <= 0 {
		return fmt.Errorf("%w: order list limit is negative or zero", errs.ErrIncorrectData)
	}

	return nil
}

type PaginationCursor struct {
	CreatedAt time.Time
	OrderUUID string
}

func (c PaginationCursor) Encode() string {
	if c.CreatedAt.IsZero() && c.OrderUUID == "" {
		return ""
	}

	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(data)
}

func DecodeTokenToCursor(token string) (PaginationCursor, error) {
	if token == "" {
		return PaginationCursor{}, nil
	}

	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return PaginationCursor{}, err
	}

	var c PaginationCursor
	err = json.Unmarshal(data, &c)

	return c, err
}
