package order

import (
	"fmt"

	"github.com/nullableocean/grpcservices/order/service"
)

var (
	ErrNotAllowedMarket = fmt.Errorf("%w: market not allowed", service.ErrInvalidData)
)
