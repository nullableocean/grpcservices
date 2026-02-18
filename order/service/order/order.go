package order

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/pkg/order"
	"github.com/nullableocean/grpcservices/pkg/roles"
)

type SpotInstrument interface {
	ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error)
}

type UserService interface {
	GetUser(ctx context.Context, id int64) (*domain.User, error)
}

type OrderService struct {
	spotInstrument SpotInstrument
	UserService    UserService

	store  map[int64]*domain.Order
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewOrderService(spotInstrument SpotInstrument, userService UserService) *OrderService {
	return &OrderService{
		spotInstrument: spotInstrument,
		UserService:    userService,
		store:          make(map[int64]*domain.Order),
		mu:             sync.RWMutex{},
	}
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderId int64, userId int64) (order.OrderStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, found := s.store[orderId]
	if !found {
		return 0, fmt.Errorf("%w: order not found. id: %d", service.ErrNotFound, orderId)
	}

	if order.UserId() != userId {
		return 0, fmt.Errorf("%w: invalid userid. order_id: %d, user_id: %d", service.ErrInvalidData, orderId, userId)
	}

	return order.Status(), nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userId int64, data domain.CreateOrderDto) (*domain.Order, error) {
	if err := s.validateCreateOrderData(data); err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	allowedMarkets, err := s.spotInstrument.ViewMarkets(ctx, user.Roles())
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
		return fmt.Errorf("%w: create order: incorrect order type value", service.ErrInvalidData)
	}

	if dto.Price == 0 {
		return fmt.Errorf("%w: create order: incorrect price value", service.ErrInvalidData)
	}

	if dto.Quantity == 0 {
		return fmt.Errorf("%w: create order: incorrect quantity value", service.ErrInvalidData)
	}

	return nil
}
