package order

import (
	"context"
	"fmt"

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

type OrderStore interface {
	Get(ctx context.Context, id int64) (*domain.Order, error)
	Create(ctx context.Context, orderData *domain.CreateOrderDto) (*domain.Order, error)
	UpdateStatus(ctx context.Context, order *domain.Order, newStatus order.OrderStatus) error
}

type OrderStatusApprover interface {
	CanChangeStatus(ctx context.Context, order *domain.Order, newStatus order.OrderStatus) error
}

type OrderService struct {
	spotInstrument SpotInstrument
	userService    UserService
	statusApprover OrderStatusApprover

	store OrderStore
}

func NewOrderService(store OrderStore, spotInstrument SpotInstrument, userService UserService, approver OrderStatusApprover) *OrderService {
	return &OrderService{
		spotInstrument: spotInstrument,
		userService:    userService,
		statusApprover: approver,
		store:          store,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, orderData *domain.CreateOrderDto) (*domain.Order, error) {
	if err := s.validateCreateOrderData(orderData); err != nil {
		return nil, err
	}

	user, err := s.userService.GetUser(ctx, orderData.UserId)
	if err != nil {
		return nil, err
	}

	allowedMarkets, err := s.spotInstrument.ViewMarkets(ctx, user.Roles())
	if err != nil {
		return nil, err
	}

	marketId := orderData.MarketId
	ok := false
	for _, market := range allowedMarkets {
		if marketId == market.Id() {
			ok = true
			break
		}
	}

	if !ok {
		return nil, fmt.Errorf("%w:market_id: %d", ErrNotAllowedMarket, marketId)
	}

	newOrder, err := s.store.Create(ctx, orderData)
	if err != nil {
		return nil, fmt.Errorf("create order error: %w", err)
	}

	return newOrder, nil
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderId int64, userId int64) (order.OrderStatus, error) {
	order, err := s.store.Get(ctx, orderId)
	if err != nil {
		return 0, fmt.Errorf("get order error: %w", service.ErrNotFound)
	}

	if order.UserId() != userId {
		return 0, fmt.Errorf("%w:invalid userid. order_id: %d, user_id: %d", service.ErrInvalidData, orderId, userId)
	}

	return order.Status(), nil
}

func (s *OrderService) ChangeStatus(ctx context.Context, orderId int64, newStatus order.OrderStatus) (order.OrderStatus, error) {
	order, err := s.store.Get(ctx, orderId)
	if err != nil {
		return 0, fmt.Errorf("get order error: %w", service.ErrNotFound)
	}

	err = s.statusApprover.CanChangeStatus(ctx, order, newStatus)
	if err != nil {
		return 0, err
	}

	err = s.store.UpdateStatus(ctx, order, newStatus)
	if err != nil {
		return 0, err
	}

	return newStatus, nil
}

func (s *OrderService) validateCreateOrderData(dto *domain.CreateOrderDto) error {
	if dto.UserId < 0 {
		return fmt.Errorf("%w: create order: invalid user id", service.ErrInvalidData)
	}

	if dto.OrderType <= 0 {
		return fmt.Errorf("%w: create order: invalid order type value", service.ErrInvalidData)
	}

	if dto.Price <= 0 {
		return fmt.Errorf("%w: create order: invalid price value", service.ErrInvalidData)
	}

	if dto.Quantity <= 0 {
		return fmt.Errorf("%w: create order: invalid quantity value", service.ErrInvalidData)
	}

	return nil
}
