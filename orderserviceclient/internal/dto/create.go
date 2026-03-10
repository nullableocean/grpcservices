package dto

import (
	"fmt"

	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/shopspring/decimal"
)

type CreateOrderDto struct {
	OrderType  order.OrderType
	UserUuid   string
	MarketUuid string
	Price      decimal.Decimal
	Quantity   decimal.Decimal
}

func (d *CreateOrderDto) Validate() error {
	if d.OrderType <= 0 {
		return fmt.Errorf("order type undefined")
	}
	if d.UserUuid == "" {
		return fmt.Errorf("empty user uuid")
	}
	if d.MarketUuid == "" {
		return fmt.Errorf("empty market uuid")
	}
	if d.Price.IsNegative() {
		return fmt.Errorf("negative price")
	}
	if d.Quantity.IsNegative() {
		return fmt.Errorf("negative quantity")
	}

	return nil
}

type StreamOrderUpdateDto struct {
	OrderUuid string
	UserUuid  string
}

func (d *StreamOrderUpdateDto) Validate() error {
	if d.UserUuid == "" {
		return fmt.Errorf("empty user uuid")
	}
	if d.OrderUuid == "" {
		return fmt.Errorf("empty order uuid")
	}

	return nil
}
