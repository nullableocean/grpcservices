package order

import (
	"context"
	"fmt"
	"sync"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/stockmarket"
	"github.com/nullableocean/grpcservices/pkg/order"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"go.uber.org/zap"
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

type RoleAccess interface {
	CanCreate(user *domain.User, orderType order.OrderType) bool
}

type SpotMarket interface {
	CanCreate(user *domain.User, orderType order.OrderType) bool
}

type OrderService struct {
	spotMarket *stockmarket.StockMarketService

	spotInstrument SpotInstrument
	userService    UserService
	roleAccesser   RoleAccess
	statusApprover OrderStatusApprover

	store OrderStore

	updatesSub map[int64]map[int]*innersub
	nextSubId  int

	mu sync.RWMutex

	logger *zap.Logger
}

func NewOrderService(logger *zap.Logger, spotMarket *stockmarket.StockMarketService, store OrderStore, spotInstrument SpotInstrument, userService UserService, marketAccess RoleAccess) *OrderService {
	s := &OrderService{
		spotMarket: spotMarket,

		spotInstrument: spotInstrument,
		userService:    userService,
		roleAccesser:   marketAccess,
		store:          store,

		statusApprover: &StatusApprover{},

		updatesSub: make(map[int64]map[int]*innersub),
		mu:         sync.RWMutex{},

		logger: logger,
	}

	s.listenMarketUpdates()
	return s
}

func (s *OrderService) CreateOrder(ctx context.Context, orderData *domain.CreateOrderDto) (*domain.Order, error) {
	if err := s.validateCreateOrderData(orderData); err != nil {
		return nil, err
	}

	user, err := s.userService.GetUser(ctx, orderData.UserId)
	if err != nil {
		return nil, err
	}

	if !s.roleAccesser.CanCreate(user, orderData.OrderType) {
		return nil, fmt.Errorf("%w: user havent permission for create this order", ErrNotAllowed)
	}

	allowedMarkets, err := s.spotInstrument.ViewMarkets(ctx, user.Roles())
	if err != nil {
		return nil, err
	}

	marketId := orderData.MarketId
	ok := false
	for _, market := range allowedMarkets {
		if marketId == market.Id {
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

	err = s.spotMarket.SendOrder(ctx, newOrder)

	return newOrder, err
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

func (s *OrderService) SubOrderUpdates(ctx context.Context, orderId int64, userId int64) (*Sub, error) {
	var out chan order.OrderStatus

	_, err := s.GetOrderStatus(ctx, orderId, userId)
	if err != nil {
		return nil, err
	}

	out = make(chan order.OrderStatus)
	s.mu.Lock()
	s.nextSubId++
	sub := &Sub{
		Id:       s.nextSubId,
		StatusCh: out,
	}

	if s.updatesSub[orderId] == nil {
		s.updatesSub[orderId] = make(map[int]*innersub)
	}

	s.updatesSub[orderId][sub.Id] = &innersub{statusCh: out}
	s.mu.Unlock()

	return sub, nil
}

func (s *OrderService) DissubOrderUpdates(ctx context.Context, orderId int64, subId int) {
	s.mu.Lock()

	delete(s.updatesSub[orderId], subId)
	s.mu.Unlock()
}

func (s *OrderService) listenMarketUpdates() {
	sub := s.spotMarket.Updates()

	go func() {
		ctx := context.Background()
		for e := range sub.UpCh {
			updatedStatus, err := s.ChangeStatus(ctx, e.OrderId, e.NewStatus)
			if err != nil {
				s.logger.Warn("error update status in order",
					zap.Int64("order_id", e.OrderId),
					zap.String("new_status", order.MapOrderStatusToString(e.NewStatus)),
				)

				s.closeSubs(e.OrderId)
				continue
			}

			s.sendUpdatedStatus(e.OrderId, updatedStatus)
		}
	}()
}

func (s *OrderService) sendUpdatedStatus(orderId int64, newStatus order.OrderStatus) {
	s.mu.RLock()

	for _, sub := range s.updatesSub[orderId] {
		select {
		case <-sub.close:
		case sub.statusCh <- newStatus:
		}
	}

	s.mu.RUnlock()
}

func (s *OrderService) closeSubs(orderId int64) {
	s.mu.Lock()

	for _, sub := range s.updatesSub[orderId] {
		close(sub.close)
		close(sub.statusCh)
	}

	delete(s.updatesSub, orderId)
	s.mu.Lock()
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
