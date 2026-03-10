package client

import "github.com/nullableocean/grpcservices/shared/order"

type Response struct {
	NewOrderUuid string
	Status       order.OrderStatus
}

type StreamData struct {
	NewStatus order.OrderStatus
	Err       error
}
