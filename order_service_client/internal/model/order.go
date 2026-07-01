package model

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	UUID       string
	MarketUUID string
	Type       OrderType
	Status     OrderStatus
	Side       OrderSide
	Price      decimal.Decimal
	Quantity   decimal.Decimal
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (o *Order) String() string {
	return fmt.Sprintf("UUID: %s\nMUUID: %s\nSide: %s\nStatus: %s\nType: %s\nPrice: %s\nQuantity: %s\nCreated: %s\nUpdated: %s\n",
		o.UUID,
		o.MarketUUID,
		o.Side,
		o.Status,
		o.Type,
		o.Price.String(),
		o.Quantity.String(),
		o.CreatedAt.Format(time.DateTime),
		o.UpdatedAt.Format(time.DateTime),
	)
}
