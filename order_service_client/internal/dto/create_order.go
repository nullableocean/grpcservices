package dto

import (
	"fmt"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"github.com/shopspring/decimal"
)

type CreateOrderParameters struct {
	UserUUID   string
	MarketUUID string
	Price      decimal.Decimal
	Quantity   decimal.Decimal
	Type       model.OrderType
	Side       model.OrderSide
}

func (d *CreateOrderParameters) Validate() error {
	if d.UserUUID == "" {
		return fmt.Errorf("empty user uuid")
	}

	if d.MarketUUID == "" {
		return fmt.Errorf("empty market uuid")
	}

	if !d.Type.IsValid() {
		return fmt.Errorf("undefined order type")
	}

	if !d.Side.IsValid() {
		return fmt.Errorf("undefined order side")
	}

	if d.Type != model.OrderTypeMarket && d.Price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("price must be positive for limit/stop orders")
	}

	if d.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("quantity must be positive")
	}

	return nil
}
