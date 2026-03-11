package order

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/shared/roles"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type SpotInstrument interface {
	ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error)
}

type UserService interface {
	GetUser(ctx context.Context, userUuid string) (*domain.User, error)
}

type OrderStore interface {
	Get(ctx context.Context, id string) (*domain.Order, error)
	Save(ctx context.Context, ord *domain.Order) error
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, e inside.Event)
}

type RoleInspector interface {
	CanCreate(user *domain.User, orderType order.OrderType) bool
}

type OrderService struct {
	spotInstrument  SpotInstrument
	userService     UserService
	roleInspect     RoleInspector
	eventDispatcher EventDispatcher

	store  OrderStore
	logger *zap.Logger
}

func NewOrderService(
	logger *zap.Logger,
	store OrderStore,
	spotInstrument SpotInstrument,
	userService UserService,
	eventDispatcher EventDispatcher,
	rInspect RoleInspector) *OrderService {

	return &OrderService{
		spotInstrument:  spotInstrument,
		userService:     userService,
		roleInspect:     rInspect,
		store:           store,
		eventDispatcher: eventDispatcher,

		logger: logger,
	}
}

func (s *OrderService) ChangeStatus(ctx context.Context, orderUuid string, newStatus order.OrderStatus) (order.OrderStatus, error) {
	o, err := s.store.Get(ctx, orderUuid)
	if err != nil {
		return 0, fmt.Errorf("get order error: %w", errs.ErrNotFound)
	}

	allowedStatuses := order.AllowedTransitions(o.GetStatus())
	if !slices.Contains(allowedStatuses, newStatus) {
		return 0, errs.ErrStatusUnavailable
	}

	o.Status = newStatus
	err = s.store.Save(ctx, o)
	if err != nil {
		return 0, err
	}

	updatedAt := time.Now()
	s.eventDispatcher.Dispatch(ctx, &inside.NewStatusEvent{
		OrderUuid: orderUuid,
		NewStatus: newStatus,
		UpdatedAt: updatedAt,
	})

	return newStatus, nil
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderUuid string, userUuid string) (order.OrderStatus, error) {
	o, err := s.FindOrderForUser(ctx, orderUuid, userUuid)
	if err != nil {
		return 0, err
	}

	return o.GetStatus(), nil
}

func (s *OrderService) FindOrderForUser(ctx context.Context, orderUuid string, userUuid string) (*domain.Order, error) {
	o, err := s.FindOrder(ctx, orderUuid)
	if err != nil {
		return nil, err
	}

	if o.GetUserUuid() != userUuid {
		return nil, fmt.Errorf("%w:invalid userid. order_id: %s, user_id: %s", errs.ErrInvalidData, orderUuid, userUuid)
	}

	return o, nil
}

func (s *OrderService) FindOrder(ctx context.Context, orderUuid string) (*domain.Order, error) {
	o, err := s.store.Get(ctx, orderUuid)
	if err != nil {
		return nil, fmt.Errorf("get order error: %w", errs.ErrNotFound)
	}

	return o, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, orderData *dto.CreateOrderDto) (*domain.Order, error) {
	ctx, span := otel.Tracer("order_service").Start(ctx, "create_order")
	defer span.End()

	if err := orderData.Validate(); err != nil {
		return nil, err
	}

	s.logger.Info("get user", zap.String("user_uuid", orderData.UserUuid))
	user, err := s.userService.GetUser(ctx, orderData.UserUuid)
	if err != nil {
		span.AddEvent("failed get user")
		return nil, err
	}

	if !s.roleInspect.CanCreate(user, orderData.OrderType) {
		s.logger.Info(
			"user havent access for this order type",
			zap.String("user_uuid", orderData.UserUuid),
			zap.String("type", orderData.OrderType.String()),
		)
		span.AddEvent("user havent permission")

		return nil, fmt.Errorf("%w: user havent permission for create this order", errs.ErrNotAllowed)
	}

	s.logger.Info("get allowed markets")
	allowedMarkets, err := s.spotInstrument.ViewMarkets(ctx, user.GetRoles())
	if err != nil {
		span.AddEvent("failed get markets")

		return nil, err
	}

	ok := false
	for _, allowedMarket := range allowedMarkets {
		if orderData.MarketUuid == allowedMarket.UUID {
			ok = true
			break
		}
	}
	if !ok {
		s.logger.Info("market not allowed for user")
		span.AddEvent("market not allowed")

		return nil, fmt.Errorf("%w:market_uuid: %s", errs.ErrNotAllowedMarket, orderData.MarketUuid)
	}

	createdAt := time.Now()
	newOrder := &domain.Order{
		UUID:       uuid.NewString(),
		UserUuid:   orderData.UserUuid,
		MarketUuid: orderData.MarketUuid,
		Price:      orderData.Price,
		Quantity:   orderData.Quantity,
		OrderType:  orderData.OrderType,
		Status:     order.ORDER_STATUS_CREATED,
		CreatedAt:  createdAt,
	}

	s.logger.Info("store order")
	err = s.store.Save(ctx, newOrder)
	if err != nil {
		span.AddEvent("failed store order")

		return nil, fmt.Errorf("store order error: %w", err)
	}

	s.logger.Info("dispatch created event")
	s.eventDispatcher.Dispatch(ctx, &inside.OrderCreatedEvent{
		Order: newOrder,
	})

	return newOrder, err
}
