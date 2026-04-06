package spotinstrument

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockMarketRepository struct {
	mock.Mock
}

func (m *mockMarketRepository) FindEnabledByRoles(ctx context.Context, roles []model.UserRole) ([]*model.Market, error) {
	args := m.Called(ctx, roles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Market), args.Error(1)
}

func (m *mockMarketRepository) FindByUUID(ctx context.Context, uuid string) (*model.Market, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Market), args.Error(1)
}

func (m *mockMarketRepository) Create(ctx context.Context, market *model.Market) error {
	args := m.Called(ctx, market)
	return args.Error(0)
}

func (m *mockMarketRepository) Update(ctx context.Context, market *model.Market) error {
	args := m.Called(ctx, market)
	return args.Error(0)
}

func (m *mockMarketRepository) Delete(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

type mockSpotInstrumentMetrics struct {
	mock.Mock
}

func (m *mockSpotInstrumentMetrics) ViewMarkets(ctx context.Context) {
	m.Called(ctx)
}

func (m *mockSpotInstrumentMetrics) FailedViewMarkets(ctx context.Context) {
	m.Called(ctx)
}

func (m *mockSpotInstrumentMetrics) FailedFindMarket(ctx context.Context) {
	m.Called(ctx)
}

// --- Helpers ---

func newTestMarket(uuid, name string, enabled bool, roles []model.UserRole) *model.Market {
	var deletedAt *time.Time
	return &model.Market{
		UUID:         uuid,
		Name:         name,
		IsEnabled:    enabled,
		DeletedAt:    deletedAt,
		AllowedRoles: roles,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestSpotInstrument_ViewMarkets(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	marketRepo := new(mockMarketRepository)
	metrics := new(mockSpotInstrumentMetrics)

	svc := NewSpotInstrument(logger, marketRepo, metrics)

	t.Run("successful retrieval", func(t *testing.T) {
		userRoles := []model.UserRole{model.UserRoleTrader}
		expectedMarkets := []*model.Market{
			newTestMarket("market12-test-uuid-test-marketmarket", "BTC/USDT", true, []model.UserRole{model.UserRoleTrader}),
			newTestMarket("market12-test-uuid-test-marketmarket", "ETH/USDT", true, nil),
		}

		marketRepo.On("FindEnabledByRoles", mock.Anything, userRoles).Return(expectedMarkets, nil).Once()
		metrics.On("ViewMarkets", mock.Anything).Return().Once()

		markets, err := svc.ViewMarkets(ctx, userRoles)

		require.NoError(t, err)
		assert.Equal(t, expectedMarkets, markets)

		marketRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		userRoles := []model.UserRole{model.UserRoleTrader}
		repoErr := errors.New("database connection failed")

		marketRepo.On("FindEnabledByRoles", mock.Anything, userRoles).Return(nil, repoErr).Once()
		metrics.On("ViewMarkets", mock.Anything).Return().Once()
		metrics.On("FailedViewMarkets", mock.Anything).Return().Once()

		markets, err := svc.ViewMarkets(ctx, userRoles)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to get markets")
		assert.Nil(t, markets)

		marketRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("empty roles should be passed to repository", func(t *testing.T) {
		var userRoles []model.UserRole
		expectedMarkets := []*model.Market{
			newTestMarket("public12-test-uuid-test-marketmarket", "BTC/USDT", true, nil),
		}

		marketRepo.On("FindEnabledByRoles", mock.Anything, userRoles).Return(expectedMarkets, nil).Once()
		metrics.On("ViewMarkets", mock.Anything).Return().Once()

		markets, err := svc.ViewMarkets(ctx, userRoles)

		require.NoError(t, err)
		assert.Equal(t, expectedMarkets, markets)

		marketRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})

	t.Run("nil roles slice", func(t *testing.T) {
		var userRoles []model.UserRole = nil
		expectedMarkets := []*model.Market{
			newTestMarket("public12-test-uuid-test-marketmarket", "BTC/USDT", true, nil),
		}

		marketRepo.On("FindEnabledByRoles", mock.Anything, userRoles).Return(expectedMarkets, nil).Once()
		metrics.On("ViewMarkets", mock.Anything).Return().Once()

		markets, err := svc.ViewMarkets(ctx, userRoles)

		require.NoError(t, err)
		assert.Equal(t, expectedMarkets, markets)

		marketRepo.AssertExpectations(t)
		metrics.AssertExpectations(t)
	})
}
