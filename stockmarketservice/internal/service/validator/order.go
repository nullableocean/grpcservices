package validator

import (
	"fmt"

	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/errs"
)

func ValidateOrder(o *domain.Order) error {
	if o.UUID == "" {
		return fmt.Errorf("%w: empty UUID", errs.ErrInvalidData)
	}

	if o.MarketUuid == "" {
		return fmt.Errorf("%w: empty MarketUuid", errs.ErrInvalidData)
	}

	if o.UserUuid == "" {
		return fmt.Errorf("%w: empty UserUuid", errs.ErrInvalidData)
	}

	if o.Quantity <= 0 {
		return fmt.Errorf("%w: invalid quantity", errs.ErrInvalidData)
	}

	if o.Price.Decimal.IsNegative() {
		return fmt.Errorf("%w: negative price", errs.ErrInvalidData)
	}

	if o.OrderType.String() == "" {
		return fmt.Errorf("%w: invalid order type", errs.ErrInvalidData)
	}

	return nil
}
