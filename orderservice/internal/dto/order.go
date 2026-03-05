package dto

import (
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/shared/money"
	"github.com/nullableocean/grpcservices/shared/order"
)

type CreateOrderDto struct {
	UserUuid   string
	MarketUuid string
	Price      money.Money
	Quantity   int64
	OrderType  order.OrderType
}

func (dto *CreateOrderDto) Validate() error {
	if dto.UserUuid == "" {
		return fmt.Errorf("%w: create order: empty user uuid", errs.ErrInvalidData)
	}

	if dto.MarketUuid == "" {
		return fmt.Errorf("%w: create order: empty market uuid", errs.ErrInvalidData)
	}

	if dto.OrderType <= 0 {
		return fmt.Errorf("%w: create order: invalid order type value", errs.ErrInvalidData)
	}

	if dto.Price.Decimal.IsNegative() {
		return fmt.Errorf("%w: create order: invalid price value", errs.ErrInvalidData)
	}

	if dto.Quantity <= 0 {
		return fmt.Errorf("%w: create order: invalid quantity value", errs.ErrInvalidData)
	}

	return nil
}
