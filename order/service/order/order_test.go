package order

import (
	"context"
	"testing"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/auth"
	"github.com/nullableocean/grpcservices/pkg/order"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

type mockOrderStore struct {
	mock.Mock
}

func (m *mockOrderStore) Get(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderStore) Create(ctx context.Context, orderData *domain.CreateOrderDto) (*domain.Order, error) {
	args := m.Called(ctx, orderData)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderStore) UpdateStatus(ctx context.Context, order *domain.Order, newStatus order.OrderStatus) error {
	args := m.Called(ctx, order, newStatus)
	order.SetStatus(newStatus)
	return args.Error(0)
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()

	spotInstrument := &mockSpotInstrument{}
	userService := &mockUserService{}
	orderStore := &mockOrderStore{}
	roleAccess := auth.NewRoleAccessService()

	orderServ := NewOrderService(orderStore, spotInstrument, userService, roleAccess)

	passSer := auth.PasswordService{}
	hash, _ := passSer.GetHashForPassword("password")
	user := domain.NewUser(&domain.CreateUserDto{
		Id:       1,
		Username: "testuser",
		PassHash: hash,
		Roles:    []roles.UserRole{roles.USER_VERIFIED},
	})

	market := domain.NewMarket(1, "BTC/USDT")

	orderData := &domain.CreateOrderDto{
		UserId:    1,
		MarketId:  1,
		Price:     50000.0,
		Quantity:  1,
		OrderType: order.ORDER_TYPE_BUY,
	}

	expectedOrder := domain.NewOrder(1, orderData)

	userService.On("GetUser", ctx, int64(1)).Return(user, nil)
	spotInstrument.On("ViewMarkets", ctx, user.Roles()).Return([]*domain.Market{market}, nil)
	orderStore.On("Create", ctx, orderData).Return(expectedOrder, nil)

	result, err := orderServ.CreateOrder(ctx, orderData)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, result)

	userService.AssertExpectations(t)
	spotInstrument.AssertExpectations(t)
	orderStore.AssertExpectations(t)
}

func TestOrderService_CreateOrderWithRoleRestrictions(t *testing.T) {
	ctx := context.Background()

	spotInstrument := &mockSpotInstrument{}
	userService := &mockUserService{}
	orderStore := &mockOrderStore{}
	roleAccess := auth.NewRoleAccessService()

	orderServ := NewOrderService(orderStore, spotInstrument, userService, roleAccess)

	passSer := auth.PasswordService{}
	hash, _ := passSer.GetHashForPassword("password")

	btcMarket := domain.NewMarket(1, "BTC/USDT")
	allowedMarkets := []*domain.Market{btcMarket}

	t.Run("guest havent acces to but/sell", func(t *testing.T) {
		guestUser := domain.NewUser(&domain.CreateUserDto{
			Id:       1,
			Username: "guestuser",
			PassHash: hash,
			Roles:    []roles.UserRole{roles.USER_GUEST},
		})

		buyOrderData := &domain.CreateOrderDto{
			UserId:    1,
			MarketId:  1,
			Price:     50000.0,
			Quantity:  1,
			OrderType: order.ORDER_TYPE_BUY,
		}

		userService.On("GetUser", ctx, int64(1)).Return(guestUser, nil)
		spotInstrument.On("ViewMarkets", ctx, guestUser.Roles()).Return(allowedMarkets, nil)

		res, err := orderServ.CreateOrder(ctx, buyOrderData)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNotAllowed)

		userService.On("GetUser", ctx, int64(1)).Return(guestUser, nil)
		spotInstrument.On("ViewMarkets", ctx, guestUser.Roles()).Return(allowedMarkets, nil)

		sellOrderData := &domain.CreateOrderDto{
			UserId:    1,
			MarketId:  1,
			Price:     50000.0,
			Quantity:  1,
			OrderType: order.ORDER_TYPE_SELL,
		}

		res2, err2 := orderServ.CreateOrder(ctx, sellOrderData)
		assert.Error(t, err2)
		assert.Nil(t, res2)
		assert.ErrorIs(t, err2, ErrNotAllowed)
	})

	t.Run("verified user has access to buy", func(t *testing.T) {
		verifiedUser := domain.NewUser(&domain.CreateUserDto{
			Id:       2,
			Username: "verifieduser",
			PassHash: hash,
			Roles:    []roles.UserRole{roles.USER_VERIFIED},
		})

		verifiedOrderData := &domain.CreateOrderDto{
			UserId:    2,
			MarketId:  1,
			Price:     50000.0,
			Quantity:  1,
			OrderType: order.ORDER_TYPE_BUY,
		}

		userService.On("GetUser", ctx, int64(2)).Return(verifiedUser, nil)
		spotInstrument.On("ViewMarkets", ctx, verifiedUser.Roles()).Return(allowedMarkets, nil)
		orderStore.On("Create", ctx, verifiedOrderData).Return(domain.NewOrder(2, verifiedOrderData), nil)

		res, err := orderServ.CreateOrder(ctx, verifiedOrderData)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("seller user has access to sell", func(t *testing.T) {

		sellerUser := domain.NewUser(&domain.CreateUserDto{
			Id:       3,
			Username: "selleruser",
			PassHash: hash,
			Roles:    []roles.UserRole{roles.USER_SELLER},
		})

		sellerOrderData := &domain.CreateOrderDto{
			UserId:    3,
			MarketId:  1,
			Price:     50000.0,
			Quantity:  1,
			OrderType: order.ORDER_TYPE_SELL,
		}

		userService.On("GetUser", ctx, int64(3)).Return(sellerUser, nil)
		spotInstrument.On("ViewMarkets", ctx, sellerUser.Roles()).Return(allowedMarkets, nil)
		orderStore.On("Create", ctx, sellerOrderData).Return(domain.NewOrder(3, sellerOrderData), nil)

		res, err := orderServ.CreateOrder(ctx, sellerOrderData)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestOrderService_CreateOrderWithNotAllowedMarket(t *testing.T) {
	ctx := context.Background()

	spotInstrument := &mockSpotInstrument{}
	userService := &mockUserService{}
	orderStore := &mockOrderStore{}
	roleAccess := auth.NewRoleAccessService()

	orderServ := NewOrderService(orderStore, spotInstrument, userService, roleAccess)

	passSer := auth.PasswordService{}
	hash, _ := passSer.GetHashForPassword("password")
	user := domain.NewUser(&domain.CreateUserDto{
		Id:       1,
		Username: "testuser",
		PassHash: hash,
		Roles:    []roles.UserRole{roles.USER_VERIFIED},
	})

	allowedMarket := domain.NewMarket(2, "ETH/USDT")

	orderData := &domain.CreateOrderDto{
		UserId:    1,
		MarketId:  1, // not allowed/not existed/not enabled
		Price:     50000.0,
		Quantity:  1,
		OrderType: order.ORDER_TYPE_BUY,
	}

	userService.On("GetUser", ctx, int64(1)).Return(user, nil)
	spotInstrument.On("ViewMarkets", ctx, user.Roles()).Return([]*domain.Market{allowedMarket}, nil)

	result, err := orderServ.CreateOrder(ctx, orderData)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotAllowedMarket)
	assert.ErrorIs(t, err, service.ErrInvalidData)

	userService.AssertExpectations(t)
	spotInstrument.AssertExpectations(t)
}

func TestOrderService_GetOrderStatus(t *testing.T) {
	ctx := context.Background()

	spotInstrument := &mockSpotInstrument{}
	userService := &mockUserService{}
	orderStore := &mockOrderStore{}
	roleAccess := auth.NewRoleAccessService()

	orderServ := NewOrderService(orderStore, spotInstrument, userService, roleAccess)

	orderData := &domain.CreateOrderDto{
		UserId:    1,
		MarketId:  1,
		Price:     50000.0,
		Quantity:  1,
		OrderType: order.ORDER_TYPE_BUY,
	}
	expectedOrder := domain.NewOrder(1, orderData)

	orderStore.On("Get", ctx, int64(1)).Return(expectedOrder, nil)
	status, err := orderServ.GetOrderStatus(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, order.ORDER_STATUS_CREATED, status)

	orderStore.AssertExpectations(t)
}

func TestOrderService_ChangeStatus(t *testing.T) {
	ctx := context.Background()

	spotInstrument := &mockSpotInstrument{}
	userService := &mockUserService{}
	orderStore := &mockOrderStore{}
	roleAccess := auth.NewRoleAccessService()

	orderServ := NewOrderService(orderStore, spotInstrument, userService, roleAccess)

	passSer := auth.PasswordService{}
	hash, _ := passSer.GetHashForPassword("password")
	user := domain.NewUser(&domain.CreateUserDto{
		Id:       1,
		Username: "testuser",
		PassHash: hash,
		Roles:    []roles.UserRole{roles.USER_VERIFIED},
	})

	orderData := &domain.CreateOrderDto{
		UserId:    1,
		MarketId:  1,
		Price:     50000.0,
		Quantity:  1,
		OrderType: order.ORDER_TYPE_BUY,
	}
	expectedOrder := domain.NewOrder(1, orderData)
	newStatus := order.ORDER_STATUS_PENDING

	userService.On("GetUser", ctx, int64(1)).Return(user, nil)
	orderStore.On("Get", ctx, int64(1)).Return(expectedOrder, nil)
	orderStore.On("UpdateStatus", ctx, expectedOrder, newStatus).Return(nil)

	t.Run("valid status change", func(t *testing.T) {
		resultStatus, err := orderServ.ChangeStatus(ctx, 1, newStatus)

		assert.NoError(t, err)
		assert.Equal(t, newStatus, resultStatus)

		status, err := orderServ.GetOrderStatus(ctx, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, newStatus, status)

		orderStore.AssertExpectations(t)
	})

	t.Run("not allowed status change", func(t *testing.T) {
		// pending --- > created  [X] err
		res, err := orderServ.ChangeStatus(ctx, 1, order.ORDER_STATUS_CREATED)

		assert.Error(t, err)
		assert.Zero(t, res)

		status, err := orderServ.GetOrderStatus(ctx, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, order.ORDER_STATUS_PENDING, status)

		orderStore.AssertExpectations(t)
	})
}
