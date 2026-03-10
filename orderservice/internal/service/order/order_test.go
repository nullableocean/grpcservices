package order

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/shared/money"
	sharedOrder "github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type MockSpotInstrument struct {
	mock.Mock
}

func (m *MockSpotInstrument) ViewMarkets(ctx context.Context, userRoles []roles.UserRole) ([]*domain.Market, error) {
	args := m.Called(ctx, userRoles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Market), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUser(ctx context.Context, userUuid string) (*domain.User, error) {
	args := m.Called(ctx, userUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockOrderStore struct {
	mock.Mock
}

func (m *MockOrderStore) Get(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderStore) Save(ctx context.Context, ord *domain.Order) error {
	args := m.Called(ctx, ord)
	return args.Error(0)
}

type MockRoleInspector struct {
	mock.Mock
}

func (m *MockRoleInspector) CanCreate(user *domain.User, orderType sharedOrder.OrderType) bool {
	args := m.Called(user, orderType)
	return args.Bool(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(ctx context.Context, e inside.Event) {
	m.Called(ctx, e)
}

type OrderServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	mockSpot      *MockSpotInstrument
	mockUserSvc   *MockUserService
	mockStore     *MockOrderStore
	mockEventDisp *MockEventDispatcher
	mockRoleInsp  *MockRoleInspector
	logger        *zap.Logger
	service       *OrderService
}

func (s *OrderServiceTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockSpot = new(MockSpotInstrument)
	s.mockUserSvc = new(MockUserService)
	s.mockStore = new(MockOrderStore)
	s.mockEventDisp = new(MockEventDispatcher)
	s.mockRoleInsp = new(MockRoleInspector)

	s.logger = zap.NewNop()

	s.service = NewOrderService(
		s.logger,
		s.mockStore,
		s.mockSpot,
		s.mockUserSvc,
		s.mockEventDisp,
		s.mockRoleInsp,
	)
}

func (s *OrderServiceTestSuite) getMoney(m int64) money.Money {
	return money.Money{
		Decimal: decimal.NewFromInt(m),
	}
}

func (s *OrderServiceTestSuite) getQuantity(q int64) int64 {
	return q
}

func TestOrderServiceSuite(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}

func (s *OrderServiceTestSuite) newTestOrder(uuidStr, userUuid string) *domain.Order {
	return &domain.Order{
		UUID:       uuidStr,
		UserUuid:   userUuid,
		MarketUuid: "TestCoin/USDT",
		Price:      s.getMoney(100),
		Quantity:   10,
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
		Status:     sharedOrder.ORDER_STATUS_CREATED,
	}
}

// ======== CHANGE STATUS
func (s *OrderServiceTestSuite) TestChangeStatus_Success() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	oldOrder := s.newTestOrder(orderUUID, userUUID)
	oldOrder.Status = sharedOrder.ORDER_STATUS_CREATED
	newStatus := sharedOrder.ORDER_STATUS_COMPLETED

	s.mockStore.On("Get", s.ctx, orderUUID).Return(oldOrder, nil).Once()
	s.mockStore.On("Save", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.UUID == orderUUID && o.Status == newStatus
	})).Return(nil).Once()

	s.mockEventDisp.On("Dispatch", s.ctx, mock.MatchedBy(func(e inside.Event) bool {
		ev, ok := e.(*inside.NewStatusEvent)
		return ok && ev.OrderUuid == orderUUID && ev.NewStatus == newStatus
	})).Return().Once()

	status, err := s.service.ChangeStatus(s.ctx, orderUUID, newStatus)
	s.NoError(err)
	s.Equal(newStatus, status)

	s.mockStore.AssertExpectations(s.T())
	s.mockEventDisp.AssertExpectations(s.T())
}

func (s *OrderServiceTestSuite) TestChangeStatus_OrderNotFound() {
	orderUUID := uuid.New().String()
	s.mockStore.On("Get", s.ctx, orderUUID).Return(nil, errors.New("not found")).Once()

	status, err := s.service.ChangeStatus(s.ctx, orderUUID, sharedOrder.ORDER_STATUS_COMPLETED)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotFound)
	s.Equal(sharedOrder.OrderStatus(0), status)
}

func (s *OrderServiceTestSuite) TestChangeStatus_InvalidTransition() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()

	changingOrder := s.newTestOrder(orderUUID, userUUID)
	changingOrder.Status = sharedOrder.ORDER_STATUS_PENDING

	invalidStatus := sharedOrder.ORDER_STATUS_CREATED

	s.mockStore.On("Get", s.ctx, orderUUID).Return(changingOrder, nil).Once()

	status, err := s.service.ChangeStatus(s.ctx, orderUUID, invalidStatus)
	s.Error(err)
	s.ErrorIs(err, errs.ErrStatusUnavailable)
	s.Equal(sharedOrder.OrderStatus(0), status)
}

func (s *OrderServiceTestSuite) TestChangeStatus_SaveError() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	oldOrder := s.newTestOrder(orderUUID, userUUID)
	oldOrder.Status = sharedOrder.ORDER_STATUS_CREATED
	newStatus := sharedOrder.ORDER_STATUS_COMPLETED

	s.mockStore.On("Get", s.ctx, orderUUID).Return(oldOrder, nil).Once()
	s.mockStore.On("Save", s.ctx, mock.Anything).Return(errors.New("db error")).Once()

	status, err := s.service.ChangeStatus(s.ctx, orderUUID, newStatus)
	s.Error(err)
	s.Equal("db error", err.Error())
	s.Equal(sharedOrder.OrderStatus(0), status)
}

// ======== GET STATUS
func (s *OrderServiceTestSuite) TestGetOrderStatus_Success() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	orderObj := s.newTestOrder(orderUUID, userUUID)
	orderObj.Status = sharedOrder.ORDER_STATUS_CREATED

	s.mockStore.On("Get", s.ctx, orderUUID).Return(orderObj, nil).Once()

	status, err := s.service.GetOrderStatus(s.ctx, orderUUID, userUUID)
	s.NoError(err)
	s.Equal(sharedOrder.ORDER_STATUS_CREATED, status)
}

func (s *OrderServiceTestSuite) TestGetOrderStatus_OrderNotFound() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()

	s.mockStore.On("Get", s.ctx, orderUUID).Return(nil, errors.New("not found")).Once()

	status, err := s.service.GetOrderStatus(s.ctx, orderUUID, userUUID)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotFound)
	s.Equal(sharedOrder.OrderStatus(0), status)
}

func (s *OrderServiceTestSuite) TestGetOrderStatus_WrongUser() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	wrongUserUUID := uuid.New().String()
	orderObj := s.newTestOrder(orderUUID, userUUID)

	s.mockStore.On("Get", s.ctx, orderUUID).Return(orderObj, nil).Once()

	status, err := s.service.GetOrderStatus(s.ctx, orderUUID, wrongUserUUID)
	s.Error(err)
	s.ErrorIs(err, errs.ErrInvalidData)
	s.Equal(sharedOrder.OrderStatus(0), status)
}

// ======== FIND ORDER
func (s *OrderServiceTestSuite) TestFindOrder_Success() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	orderObj := s.newTestOrder(orderUUID, userUUID)

	s.mockStore.On("Get", s.ctx, orderUUID).Return(orderObj, nil).Once()

	found, err := s.service.FindOrderForUser(s.ctx, orderUUID, userUUID)
	s.NoError(err)
	s.Equal(orderObj, found)
}

func (s *OrderServiceTestSuite) TestFindOrder_NotFound() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()

	s.mockStore.On("Get", s.ctx, orderUUID).Return(nil, errors.New("not found")).Once()

	found, err := s.service.FindOrderForUser(s.ctx, orderUUID, userUUID)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotFound)
	s.Nil(found)
}

func (s *OrderServiceTestSuite) TestFindOrder_WrongUser() {
	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	wrongUserUUID := uuid.New().String()
	orderObj := s.newTestOrder(orderUUID, userUUID)

	s.mockStore.On("Get", s.ctx, orderUUID).Return(orderObj, nil).Once()

	found, err := s.service.FindOrderForUser(s.ctx, orderUUID, wrongUserUUID)
	s.Error(err)
	s.ErrorIs(err, errs.ErrInvalidData)
	s.Nil(found)
}

// ======== CREATE ORDER
func (s *OrderServiceTestSuite) TestCreateOrder_Success() {
	userUUID := uuid.New().String()
	marketUUID := "market-1"
	orderType := sharedOrder.ORDER_TYPE_BUY

	price := s.getMoney(100)
	quantity := s.getQuantity(10)

	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: marketUUID,
		Price:      price,
		Quantity:   int64(quantity),
		OrderType:  orderType,
	}

	user := &domain.User{
		UUID:  userUUID,
		Roles: roles.NewRoles(roles.USER_SELLER),
	}
	markets := []*domain.Market{
		{UUID: marketUUID, Name: "BTC/USD"},
	}

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(user, nil).Once()
	s.mockRoleInsp.On("CanCreate", user, orderType).Return(true).Once()
	s.mockSpot.On("ViewMarkets", s.ctx, user.Roles.GetSlice()).Return(markets, nil).Once()
	s.mockStore.On("Save", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.UserUuid == userUUID &&
			o.MarketUuid == marketUUID &&
			o.Status == sharedOrder.ORDER_STATUS_CREATED &&
			o.UUID != ""
	})).Return(nil).Once()

	s.mockEventDisp.On("Dispatch", s.ctx, mock.MatchedBy(func(e inside.Event) bool {
		ev, ok := e.(*inside.OrderCreatedEvent)
		return ok && ev.Order.UUID != ""
	})).Return().Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.NoError(err)
	s.NotNil(order)
	s.Equal(userUUID, order.UserUuid)
	s.Equal(marketUUID, order.MarketUuid)
	s.Equal(price, order.Price)
	s.Equal(quantity, order.Quantity)
	s.Equal(orderType, order.OrderType)
	s.Equal(sharedOrder.ORDER_STATUS_CREATED, order.Status)

	s.mockUserSvc.AssertExpectations(s.T())
	s.mockRoleInsp.AssertExpectations(s.T())
	s.mockSpot.AssertExpectations(s.T())
	s.mockStore.AssertExpectations(s.T())
	s.mockEventDisp.AssertExpectations(s.T())
}

func (s *OrderServiceTestSuite) TestCreateOrder_ValidationNegativePriceError() {
	negativePrice := s.getMoney(-100)

	createDto := &dto.CreateOrderDto{
		UserUuid:   uuid.New().String(),
		MarketUuid: "market-uuid",
		Price:      negativePrice,
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.Nil(order)
}

func (s *OrderServiceTestSuite) TestCreateOrder_GetUserNotFoundError() {
	userUUID := uuid.New().String()
	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: "market-uuid",
		Price:      s.getMoney(100),
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(nil, errs.ErrNotFound).Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotFound)
	s.Nil(order)
}

func (s *OrderServiceTestSuite) TestCreateOrder_NoPermission() {
	userUUID := uuid.New().String()
	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: "market-uuid",
		Price:      s.getMoney(100),
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}
	user := &domain.User{UUID: userUUID, Roles: roles.NewRoles()}

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(user, nil).Once()
	s.mockRoleInsp.On("CanCreate", user, createDto.OrderType).Return(false).Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotAllowed)
	s.Nil(order)
}

func (s *OrderServiceTestSuite) TestCreateOrder_ViewMarketsError() {
	userUUID := uuid.New().String()
	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: "market-uuid",
		Price:      s.getMoney(100),
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}
	user := &domain.User{UUID: userUUID, Roles: roles.NewRoles(roles.USER_MODER)}

	errorMsg := "failed spot service connect"

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(user, nil).Once()
	s.mockRoleInsp.On("CanCreate", user, createDto.OrderType).Return(true).Once()
	s.mockSpot.On("ViewMarkets", s.ctx, user.Roles.GetSlice()).Return(nil, errors.New(errorMsg)).Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.Nil(order)
}

func (s *OrderServiceTestSuite) TestCreateOrder_MarketNotAllowed() {
	userUUID := uuid.New().String()
	requestedMarket := "market-uuid"
	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: requestedMarket,
		Price:      s.getMoney(100),
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}
	user := &domain.User{UUID: userUUID, Roles: roles.NewRoles(roles.USER_VERIFIED)}

	allowedMarkets := []*domain.Market{{UUID: "market-uuid-2", Name: "ETH/USD"}}

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(user, nil).Once()
	s.mockRoleInsp.On("CanCreate", user, createDto.OrderType).Return(true).Once()
	s.mockSpot.On("ViewMarkets", s.ctx, user.Roles.GetSlice()).Return(allowedMarkets, nil).Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.ErrorIs(err, errs.ErrNotAllowedMarket)
	s.Nil(order)
}

func (s *OrderServiceTestSuite) TestCreateOrder_StoreSaveError() {
	userUUID := uuid.New().String()
	requestedMarket := "market-uuid"
	createDto := &dto.CreateOrderDto{
		UserUuid:   userUUID,
		MarketUuid: requestedMarket,
		Price:      s.getMoney(100),
		Quantity:   s.getQuantity(10),
		OrderType:  sharedOrder.ORDER_TYPE_BUY,
	}
	user := &domain.User{UUID: userUUID, Roles: roles.NewRoles(roles.USER_VERIFIED)}
	markets := []*domain.Market{{UUID: requestedMarket, Name: "BTC/USD"}}

	s.mockUserSvc.On("GetUser", s.ctx, userUUID).Return(user, nil).Once()
	s.mockRoleInsp.On("CanCreate", user, createDto.OrderType).Return(true).Once()
	s.mockSpot.On("ViewMarkets", s.ctx, user.Roles.GetSlice()).Return(markets, nil).Once()

	saveError := "store cant connect to db"
	s.mockStore.On("Save", s.ctx, mock.Anything).Return(errors.New(saveError)).Once()

	order, err := s.service.CreateOrder(s.ctx, createDto)
	s.Error(err)
	s.Nil(order)
}
