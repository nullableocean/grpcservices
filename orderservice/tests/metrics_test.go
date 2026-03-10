package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside/handlers"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	"github.com/nullableocean/grpcservices/orderservice/internal/store/ram"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/server"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mocks
type mockSpotInstrument struct {
	mock.Mock
}

func (m *mockSpotInstrument) ViewMarkets(ctx context.Context, userRoles []roles.UserRole) ([]*domain.Market, error) {
	args := m.Called(ctx, userRoles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Market), args.Error(1)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetUser(ctx context.Context, userUuid string) (*domain.User, error) {
	args := m.Called(ctx, userUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockEventDispatcher struct {
	mock.Mock
}

func (m *mockEventDispatcher) Dispatch(ctx context.Context, e inside.Event) {
	m.Called(ctx, e)
}

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	userUUID := uuid.New().String()
	marketUUID := uuid.New().String()

	user := &domain.User{
		UUID:  userUUID,
		Roles: roles.NewRoles(roles.USER_SELLER),
	}

	market := &domain.Market{
		UUID: marketUUID,
		Name: "test_market",
	}

	userService := &mockUserService{}
	userService.On("GetUser", mock.Anything, userUUID).Return(user, nil)

	spotInstrument := &mockSpotInstrument{}
	spotInstrument.On("ViewMarkets", mock.Anything, user.Roles.GetSlice()).Return([]*domain.Market{market}, nil)

	eventDispatcher := &mockEventDispatcher{}
	eventDispatcher.On("Dispatch", mock.Anything, mock.MatchedBy(func(e inside.Event) bool {
		_, ok := e.(*inside.OrderCreatedEvent)
		return ok
	})).Return().Times(2)

	statusStreamer := handlers.NewStatusStreamer(logger, handlers.Option{MaxSendingProcess: 5})

	store := ram.NewOrderStore()
	roleInspector := access.NewRoleInspector()

	orderService := order.NewOrderService(
		logger,
		store,
		spotInstrument,
		userService,
		eventDispatcher,
		roleInspector,
	)

	reg := prometheus.NewRegistry()
	orderMetrics := metrics.NewOrderMetrics(reg)

	orderServer := server.NewOrderServer(logger, orderService, orderMetrics, statusStreamer)

	createReq := &orderv1.CreateOrderRequest{
		UserUuid:  userUUID,
		MarketId:  marketUUID,
		OrderType: typesv1.OrderType_ORDER_TYPE_BUY,
		Price: &typesv1.Money{
			Units: 1000,
			Nanos: 0,
		},
		Quantity: 1,
	}

	resp1, err := orderServer.CreateOrder(ctx, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.OrderUuid)

	resp2, err := orderServer.CreateOrder(ctx, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, resp2.OrderUuid)

	_, err = orderServer.GetOrderStatus(ctx, &orderv1.GetStatusRequest{
		OrderUuid: resp2.OrderUuid,
		UserUuid:  userUUID,
	})
	require.NoError(t, err)

	collectedMetrics, err := reg.Gather()
	require.NoError(t, err)

	var createCounterFound, getCounterFound, durationFound bool

	for _, mf := range collectedMetrics {
		switch mf.GetName() {
		case metrics.Namespace + "_" + metrics.CreateOrderCalls:
			createCounterFound = true
			metricsList := mf.GetMetric()
			assert.Len(t, metricsList, 1, "exactly one metric should exist for CreateOrderCalls")
			counterValue := metricsList[0].GetCounter().GetValue()
			assert.Equal(t, 2.0, counterValue, "CreateOrder counter should be 2")

		case metrics.Namespace + "_" + metrics.GetStatusCalls:
			getCounterFound = true
			metricsList := mf.GetMetric()
			assert.Len(t, metricsList, 1, "exactly one metric should exist for GetStatusCalls")
			counterValue := metricsList[0].GetCounter().GetValue()
			assert.Equal(t, 1.0, counterValue, "GetOrderStatus counter should be 1")

		case metrics.Namespace + "_" + metrics.CreateOrderDuration:
			durationFound = true
		}
	}

	assert.True(t, createCounterFound, "metric CreateOrderCalls not found")
	assert.True(t, getCounterFound, "metric GetStatusCalls not found")
	assert.True(t, durationFound, "metric CreateOrderDuration not found")

	userService.AssertExpectations(t)
	spotInstrument.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}
