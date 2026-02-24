package order

import (
	"fmt"

	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/pkg/order"
)

type innersub struct {
	statusCh chan order.OrderStatus
	close    chan struct{}
}

type Sub struct {
	Id       int
	StatusCh <-chan order.OrderStatus
}

var (
	ErrNotAllowedMarket = fmt.Errorf("%w:market not allowed for user", service.ErrInvalidData)
	ErrNotAllowed       = fmt.Errorf("%w:not allowed for user", service.ErrAccessDenied)
)
