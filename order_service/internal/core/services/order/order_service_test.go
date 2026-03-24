package order

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockOrderRepository struct {
	mock.Mock
}

func (m *mockOrderRepository) Save(ctx context.Context, order *model.Order, events ...model.Event) error {
	args := m.Called(ctx, order, events)
	return args.Error(0)
}

func (m *mockOrderRepository) Update(ctx context.Context, order *model.Order, events ...model.Event) error {
	args := m.Called(ctx, order, events)
	return args.Error(0)
}

func (m *mockOrderRepository) FindByUUID(ctx context.Context, orderUUID string) (*model.Order, error) {
	args := m.Called(ctx, orderUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

type mockSpotInstrument struct {
	mock.Mock
}

func (m *mockSpotInstrument) ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]model.Market, error) {
	args := m.Called(ctx, userRoles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Market), args.Error(1)
}

type mockAccessService struct {
	mock.Mock
}

func (m *mockAccessService) CanCreateOrder(ctx context.Context, user *model.User, params *dto.CreateOrderParameters) error {
	args := m.Called(ctx, user, params)
	return args.Error(0)
}

type mockMetricsRecorder struct {
	mock.Mock
}

func (m *mockMetricsRecorder) OrderCreated(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderCompleted(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderRejected(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderCancelled(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderFailed(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderFailedCreate(ctx context.Context) {
	m.Called(ctx)
}
func (m *mockMetricsRecorder) OrderFailedUpdate(ctx context.Context) {
	m.Called(ctx)
}

func newTestUser(roles ...model.UserRole) *model.User {
	return &model.User{
		UUID:  "test-user-uuid",
		Roles: roles,
	}
}

func newTestMarket(marketUUID string) model.Market {
	return model.Market{UUID: marketUUID}
}

func newTestOrder(uuid, userUUID, marketUUID string, side model.OrderSide, typ model.OrderType, price, quantity decimal.Decimal, status model.OrderStatus) *model.Order {
	return &model.Order{
		UUID:       uuid,
		UserUUID:   userUUID,
		MarketUUID: marketUUID,
		Side:       side,
		Type:       typ,
		Status:     status,
		Price:      price,
		Quantity:   quantity,
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	orderRepo := new(mockOrderRepository)
	spotInst := new(mockSpotInstrument)
	accessSvc := new(mockAccessService)
	metrics := new(mockMetricsRecorder)

	svc := NewOrderService(logger, orderRepo, spotInst, accessSvc, metrics)

	t.Run("success", func(t *testing.T) {
		user := newTestUser(model.UserRoleTrader)
		params := &dto.CreateOrderParameters{
			User:       user,
			MarketUUID: "BTC-USDT",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		require.NoError(t, params.Validate())

		accessSvc.On("CanCreateOrder", mock.Anything, user, params).Return(nil).Once()
		spotInst.On("ViewMarkets", mock.Anything, user.Roles).Return([]model.Market{newTestMarket("BTC-USDT")}, nil).Once()
		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*model.Order"), mock.Anything).Return(nil).Once()
		metrics.On("OrderCreated", mock.Anything).Return().Once()

		order, err := svc.CreateOrder(ctx, params)
		require.NoError(t, err)
		assert.NotEmpty(t, order.UUID)
		assert.Equal(t, user.UUID, order.UserUUID)
		assert.Equal(t, params.MarketUUID, order.MarketUUID)
		assert.Equal(t, params.Side, order.Side)
		assert.Equal(t, params.Type, order.Type)
		assert.Equal(t, model.OrderStatusCreated, order.Status)
		assert.Equal(t, params.Price, order.Price)
		assert.Equal(t, params.Quantity, order.Quantity)

		accessSvc.AssertExpectations(t)
		spotInst.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		params := &dto.CreateOrderParameters{
			User:       newTestUser(),
			MarketUUID: "",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		assert.Error(t, params.Validate())

		order, err := svc.CreateOrder(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, order)

		accessSvc.AssertNotCalled(t, "CanCreateOrder")
		spotInst.AssertNotCalled(t, "ViewMarkets")
		orderRepo.AssertNotCalled(t, "Save")
		metrics.AssertNotCalled(t, "OrderCreated")
		metrics.AssertNotCalled(t, "OrderFailedCreate")
	})

	t.Run("access denied", func(t *testing.T) {
		user := newTestUser(model.UserRoleGuest)
		params := &dto.CreateOrderParameters{
			User:       user,
			MarketUUID: "BTC-USDT",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		require.NoError(t, params.Validate())

		accessSvc.On("CanCreateOrder", mock.Anything, user, params).Return(errors.New("forbidden")).Once()

		order, err := svc.CreateOrder(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, order)

		accessSvc.AssertExpectations(t)
		spotInst.AssertNotCalled(t, "ViewMarkets")
		orderRepo.AssertNotCalled(t, "Save")
		metrics.AssertNotCalled(t, "OrderCreated")
		metrics.AssertNotCalled(t, "OrderFailedCreate")
	})

	t.Run("ViewMarkets error", func(t *testing.T) {
		user := newTestUser(model.UserRoleTrader)
		params := &dto.CreateOrderParameters{
			User:       user,
			MarketUUID: "BTC-USDT",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		require.NoError(t, params.Validate())

		accessSvc.On("CanCreateOrder", mock.Anything, user, params).Return(nil).Once()
		spotInst.On("ViewMarkets", mock.Anything, user.Roles).Return(nil, errors.New("network error")).Once()
		metrics.On("OrderFailedCreate", mock.Anything).Return().Once()

		order, err := svc.CreateOrder(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, order)

		accessSvc.AssertExpectations(t)
		spotInst.AssertExpectations(t)
		orderRepo.AssertNotCalled(t, "Save")
		metrics.AssertExpectations(t)
	})

	t.Run("market not found in allowed list", func(t *testing.T) {
		user := newTestUser(model.UserRoleTrader)
		params := &dto.CreateOrderParameters{
			User:       user,
			MarketUUID: "BTC-USDT",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		require.NoError(t, params.Validate())

		accessSvc.On("CanCreateOrder", mock.Anything, user, params).Return(nil).Once()
		spotInst.On("ViewMarkets", mock.Anything, user.Roles).Return([]model.Market{}, nil).Once()
		metrics.On("OrderFailedCreate", mock.Anything).Return().Once()

		order, err := svc.CreateOrder(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, order)

		accessSvc.AssertExpectations(t)
		spotInst.AssertExpectations(t)
		orderRepo.AssertNotCalled(t, "Save")
		metrics.AssertExpectations(t)
	})

	t.Run("repository save error", func(t *testing.T) {
		user := newTestUser(model.UserRoleTrader)
		params := &dto.CreateOrderParameters{
			User:       user,
			MarketUUID: "BTC-USDT",
			Side:       model.OrderSideBuy,
			Type:       model.OrderTypeLimit,
			Price:      decimal.NewFromInt(50000),
			Quantity:   decimal.NewFromInt(1),
		}
		require.NoError(t, params.Validate())

		accessSvc.On("CanCreateOrder", mock.Anything, user, params).Return(nil).Once()
		spotInst.On("ViewMarkets", mock.Anything, user.Roles).Return([]model.Market{newTestMarket("BTC-USDT")}, nil).Once()
		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*model.Order"), mock.Anything).Return(errors.New("db error")).Once()
		metrics.On("OrderFailedCreate", mock.Anything).Return().Once()

		order, err := svc.CreateOrder(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, order)

		accessSvc.AssertExpectations(t)
		spotInst.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})
}

func TestOrderService_GetOrder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	orderRepo := new(mockOrderRepository)
	spotInst := new(mockSpotInstrument)
	accessSvc := new(mockAccessService)
	metrics := new(mockMetricsRecorder)

	svc := NewOrderService(logger, orderRepo, spotInst, accessSvc, metrics)

	t.Run("success", func(t *testing.T) {
		orderUUID := uuid.NewString()
		userUUID := uuid.NewString()

		expectedOrder := newTestOrder(
			orderUUID,
			userUUID,
			"market12-test-uuid-test-marketmarket",
			model.OrderSideBuy,
			model.OrderTypeLimit,
			decimal.NewFromInt(1000),
			decimal.NewFromInt(2),
			model.OrderStatusCreated,
		)
		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(expectedOrder, nil).Once()

		order, err := svc.GetOrder(ctx, orderUUID, userUUID)
		require.NoError(t, err)
		assert.Equal(t, expectedOrder, order)

		orderRepo.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		orderUUID := uuid.NewString()
		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(nil, ports.ErrNotFound).Once()

		order, err := svc.GetOrder(ctx, orderUUID, "any-user")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.Nil(t, order)

		orderRepo.AssertExpectations(t)
	})

	t.Run("user does not own order", func(t *testing.T) {
		orderUUID := uuid.NewString()
		ownerUUID := uuid.NewString()
		otherUserUUID := uuid.NewString()

		order := newTestOrder(
			orderUUID,
			ownerUUID,
			"market12-test-uuid-test-marketmarket",
			model.OrderSideBuy,
			model.OrderTypeLimit,
			decimal.NewFromInt(1000),
			decimal.NewFromInt(2),
			model.OrderStatusCreated,
		)
		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(order, nil).Once()

		got, err := svc.GetOrder(ctx, orderUUID, otherUserUUID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNotFound)
		assert.Nil(t, got)

		orderRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		orderUUID := uuid.NewString()
		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(nil, errors.New("db error")).Once()

		order, err := svc.GetOrder(ctx, orderUUID, "any-user")
		assert.Error(t, err)
		assert.Nil(t, order)

		orderRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateOrder(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	orderRepo := new(mockOrderRepository)
	spotInst := new(mockSpotInstrument)
	accessSvc := new(mockAccessService)
	metrics := new(mockMetricsRecorder)

	svc := NewOrderService(logger, orderRepo, spotInst, accessSvc, metrics)

	t.Run("success update to completed", func(t *testing.T) {
		orderUUID := uuid.NewString()
		userUUID := uuid.NewString()
		newStatus := model.OrderStatusCompleted
		params := &dto.UpdateOrderParameters{Status: newStatus}
		require.NoError(t, params.Validate())

		existingOrder := newTestOrder(
			orderUUID, userUUID, "market12-test-uuid-test-marketmarket",
			model.OrderSideBuy, model.OrderTypeLimit,
			decimal.NewFromInt(1000), decimal.NewFromInt(2),
			model.OrderStatusPending,
		)

		require.True(t, existingOrder.Status.CanTransitTo(newStatus), "transition should be allowed")

		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(existingOrder, nil).Once()
		orderRepo.On("Update", mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
			return o.UUID == orderUUID && o.Status == newStatus
		}), mock.Anything).Return(nil).Once()
		metrics.On("OrderCompleted", mock.Anything).Return().Once()

		err := svc.UpdateOrder(ctx, orderUUID, params)
		assert.NoError(t, err)

		orderRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("success update to rejected", func(t *testing.T) {
		orderUUID := uuid.NewString()
		userUUID := uuid.NewString()
		newStatus := model.OrderStatusRejected
		params := &dto.UpdateOrderParameters{Status: newStatus}
		require.NoError(t, params.Validate())

		existingOrder := newTestOrder(
			orderUUID,
			userUUID,
			"market12-test-uuid-test-marketmarket",
			model.OrderSideBuy,
			model.OrderTypeLimit,
			decimal.NewFromInt(1000),
			decimal.NewFromInt(2),
			model.OrderStatusPending,
		)

		require.True(t, existingOrder.Status.CanTransitTo(newStatus), "transition should be allowed")

		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(existingOrder, nil).Once()
		orderRepo.On("Update", mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
			return o.UUID == orderUUID && o.Status == newStatus
		}), mock.Anything).Return(nil).Once()
		metrics.On("OrderRejected", mock.Anything).Return().Once()

		err := svc.UpdateOrder(ctx, orderUUID, params)
		assert.NoError(t, err)

		orderRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		params := &dto.UpdateOrderParameters{Status: "invalid"}
		assert.Error(t, params.Validate())

		err := svc.UpdateOrder(ctx, "any-uuid", params)
		assert.Error(t, err)
		orderRepo.AssertNotCalled(t, "FindByUUID")
		orderRepo.AssertNotCalled(t, "Update")
		metrics.AssertNotCalled(t, "OrderCompleted")
		metrics.AssertNotCalled(t, "OrderFailedUpdate")
	})

	t.Run("order not found", func(t *testing.T) {
		orderUUID := uuid.NewString()
		params := &dto.UpdateOrderParameters{Status: model.OrderStatusCompleted}
		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(nil, ports.ErrNotFound).Once()

		err := svc.UpdateOrder(ctx, orderUUID, params)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNotFound)

		orderRepo.AssertExpectations(t)
		orderRepo.AssertNotCalled(t, "Update")
		metrics.AssertNotCalled(t, "OrderCompleted")
		metrics.AssertNotCalled(t, "OrderFailedUpdate")
	})

	t.Run("invalid status transition", func(t *testing.T) {
		orderUUID := uuid.NewString()
		userUUID := uuid.NewString()
		newStatus := model.OrderStatusCompleted
		params := &dto.UpdateOrderParameters{Status: newStatus}

		existingOrder := newTestOrder(
			orderUUID,
			userUUID,
			"market12-test-uuid-test-marketmarket",
			model.OrderSideBuy,
			model.OrderTypeLimit,
			decimal.NewFromInt(1000),
			decimal.NewFromInt(2),
			model.OrderStatusCancelled,
		)

		require.False(t, existingOrder.Status.CanTransitTo(newStatus), "transition should be forbidden")

		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(existingOrder, nil).Once()
		metrics.On("OrderFailedUpdate", mock.Anything).Return().Once()

		err := svc.UpdateOrder(ctx, orderUUID, params)
		assert.Error(t, err)

		orderRepo.AssertExpectations(t)
		orderRepo.AssertNotCalled(t, "Update")
		metrics.AssertExpectations(t)
	})

	t.Run("repository update error", func(t *testing.T) {
		orderUUID := uuid.NewString()
		userUUID := uuid.NewString()
		newStatus := model.OrderStatusCompleted
		params := &dto.UpdateOrderParameters{Status: newStatus}

		existingOrder := newTestOrder(
			orderUUID,
			userUUID,
			"market12-test-uuid-test-marketmarket",
			model.OrderSideBuy,
			model.OrderTypeLimit,
			decimal.NewFromInt(1000),
			decimal.NewFromInt(2),
			model.OrderStatusPending,
		)

		require.True(t, existingOrder.Status.CanTransitTo(newStatus), "transition should be allowed")

		orderRepo.On("FindByUUID", mock.Anything, orderUUID).Return(existingOrder, nil).Once()
		orderRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
		metrics.On("OrderFailedUpdate", mock.Anything).Return().Once()

		err := svc.UpdateOrder(ctx, orderUUID, params)
		assert.Error(t, err)

		orderRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})
}
