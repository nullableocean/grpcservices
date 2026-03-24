package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	UUID       string          `json:"uuid"`
	UserUUID   string          `json:"user_uuid"`
	MarketUUID string          `json:"market_uuid"`
	Side       OrderSide       `json:"side"`
	Type       OrderType       `json:"type"`
	Status     OrderStatus     `json:"status"`
	Price      decimal.Decimal `json:"price"`
	Quantity   decimal.Decimal `json:"quantity"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
