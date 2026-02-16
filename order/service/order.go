package service

import (
	"context"
	"fmt"
	"main/order/client"
	"main/order/domain"
	pkg "main/pkg/order"
	"sync"
	"sync/atomic"
)

var (
	ErrNotAllowedMarket = fmt.Errorf("%w: market not allowed", ErrInvalidData)
)

type OrderService struct {
	spotClient  *client.SpotClient
	UserService *UserService

	store  map[int64]*domain.Order
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewOrderService(spotClient *client.SpotClient, userService *UserService) *OrderService {
	return &OrderService{
		spotClient:  spotClient,
		UserService: userService,
		store:       make(map[int64]*domain.Order),
		mu:          sync.RWMutex{},
	}
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderId int64, userId int64) (pkg.OrderStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, found := s.store[orderId]
	if !found {
		return 0, fmt.Errorf("%w: order not found. id: %d", ErrNotFound, orderId)
	}

	if order.UserId() != userId {
		return 0, fmt.Errorf("%w: invalid userid. order_id: %d, user_id: %d", ErrInvalidData, orderId, userId)
	}

	return order.Status(), nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userId int64, data domain.CreateOrderDto) (*domain.Order, error) {
	if err := s.validateCreateOrderData(data); err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUser(userId)
	if err != nil {
		return nil, err
	}

	allowedMarkets, err := s.spotClient.ViewMarkets(ctx, user.Roles())
	if err != nil {
		return nil, err
	}

	marketId := data.MarketId

	ok := false
	for _, market := range allowedMarkets {
		if marketId == market.Id() {
			ok = true
		}
	}

	if !ok {
		return nil, fmt.Errorf("%w: market_id: %d", ErrNotAllowedMarket, marketId)
	}

	id := s.nextId.Add(1)
	newOrder := domain.NewOrder(id, userId, data)
	s.store[id] = newOrder

	return newOrder, nil
}

func (s *OrderService) validateCreateOrderData(dto domain.CreateOrderDto) error {
	if dto.OrderType == 0 {
		return fmt.Errorf("%w: create order: incorrect order type value", ErrInvalidData)
	}

	if dto.Price == 0 {
		return fmt.Errorf("%w: create order: incorrect price value", ErrInvalidData)
	}

	if dto.Quantity == 0 {
		return fmt.Errorf("%w: create order: incorrect quantity value", ErrInvalidData)
	}

	return nil
}
