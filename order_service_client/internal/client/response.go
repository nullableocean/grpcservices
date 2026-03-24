package client

import (
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
)

type Response struct {
	NewOrderUuid string
	Status       model.OrderStatus
}

type StreamData struct {
	NewStatus model.OrderStatus
	Err       error
}
