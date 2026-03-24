package dto

import (
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/shopspring/decimal"
)

type CreateOrderParameters struct {
	User       *model.User
	MarketUUID string
	Price      decimal.Decimal
	Quantity   decimal.Decimal
	Type       model.OrderType
	Side       model.OrderSide
}

func (d *CreateOrderParameters) Validate() error {
	if d.MarketUUID == "" {
		return fmt.Errorf("%w: empty market uuid", errs.ErrIncorrectData)
	}

	if !d.Type.IsValid() {
		return fmt.Errorf("%w: undefined order type", errs.ErrIncorrectData)
	}

	if !d.Side.IsValid() {
		return fmt.Errorf("%w: undefined order side", errs.ErrIncorrectData)
	}

	if d.Type != model.OrderTypeMarket && d.Price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: price must be positive for limit/stop orders", errs.ErrIncorrectData)
	}

	if d.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: quantity must be positive", errs.ErrIncorrectData)
	}

	return nil
}
