package dto

import (
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type UpdateOrderParameters struct {
	Status model.OrderStatus
}

func (d *UpdateOrderParameters) Validate() error {
	if !d.Status.IsValid() {
		return fmt.Errorf("%w: undefined status", errs.ErrIncorrectData)
	}

	return nil
}
