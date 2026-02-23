package order

import (
	"fmt"

	"github.com/nullableocean/grpcservices/order/service"
)

var (
	ErrNotAllowedMarket = fmt.Errorf("%w:market not allowed for user", service.ErrInvalidData)
	ErrNotAllowed       = fmt.Errorf("%w:not allowed for user", service.ErrAccessDenied)
)
