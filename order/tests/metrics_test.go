package tests

import (
	"context"
	"testing"

	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/server"
	"github.com/nullableocean/grpcservices/order/service/auth"
	"github.com/nullableocean/grpcservices/order/service/metrics"
	"github.com/nullableocean/grpcservices/order/service/order"
	"github.com/nullableocean/grpcservices/order/service/store/ram"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockSpotInstrument struct {
	mock.Mock
}

func (m *mockSpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	args := m.Called(ctx, roles)
	return args.Get(0).([]*domain.Market), args.Error(1)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestMetrics(t *testing.T) {

	ctx := context.Background()

	passSer := auth.PasswordService{}
	hash, _ := passSer.GetHashForPassword("password")
	user := domain.NewUser(&domain.CreateUserDto{
		Id:       10,
		Username: "tea",
		PassHash: hash,
		Roles:    []roles.UserRole{roles.USER_SELLER},
	})

	m := domain.NewMarket(1, "test_market")

	userService := &mockUserService{}
	userService.On("GetUser", ctx, int64(10)).Return(user, nil)

	spotInstrument := &mockSpotInstrument{}
	spotInstrument.On("ViewMarkets", ctx, user.Roles()).Return([]*domain.Market{m}, nil)

	store := ram.NewOrderStore()

	orderService := order.NewOrderService(store, spotInstrument, userService)

	reg := prometheus.NewRegistry()
	orderMetrics := metrics.NewOrderMetrics(reg)
	orderServer := server.NewOrderServer(orderService, zap.NewNop(), orderMetrics)

	resp, err := orderServer.CreateOrder(context.Background(), &orderpb.CreateOrderRequest{
		UserId:    10,
		MarketId:  1,
		OrderType: 1,
		Price:     1000.0,
		Quantity:  1,
	})

	assert.NoError(t, err)

	resp, err = orderServer.CreateOrder(context.Background(), &orderpb.CreateOrderRequest{
		UserId:    10,
		MarketId:  1,
		OrderType: 1,
		Price:     1000.0,
		Quantity:  1,
	})

	assert.NoError(t, err)

	orderServer.GetOrderStatus(context.Background(), &orderpb.GetStatusRequest{
		OrderId: resp.OrderId,
		UserId:  10,
	})

	collectedMetrics, err := reg.Gather()

	assert.NoError(t, err)

	createOrdersCounter := false
	createOrdersDuration := false
	getStatusCounter := false

	for _, m := range collectedMetrics {
		if *m.Name == metrics.Namespace+"_"+metrics.CreateOrderCalls {
			createOrdersCounter = true

			collectedMetrics := m.GetMetric()

			assert.Len(t, collectedMetrics, 1)

			counterMetric := collectedMetrics[0]
			callCount := counterMetric.Counter.GetValue()

			assert.Equal(t, int(callCount), 2, "create order call count should be 2")
		}
		if *m.Name == metrics.Namespace+"_"+metrics.GetStatusCalls {
			getStatusCounter = true

			collectedMetrics := m.GetMetric()

			assert.Len(t, collectedMetrics, 1)

			counterMetric := collectedMetrics[0]
			callCount := counterMetric.Counter.GetValue()

			assert.Equal(t, int(callCount), 1, "create order call count should be 1")
		}
		if *m.Name == metrics.Namespace+"_"+metrics.CreateOrderDuration {
			createOrdersDuration = true
		}
	}

	assert.True(t, createOrdersCounter)
	assert.True(t, createOrdersDuration)
	assert.True(t, getStatusCounter)
}
