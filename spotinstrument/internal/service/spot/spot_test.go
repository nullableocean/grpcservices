package spot

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service"
	guard "github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/auth"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/store/ram"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpotService_ViewMarkets(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty slice when no markets enabled", func(t *testing.T) {
		store := ram.NewMarketStore()

		allowedRoles := []roles.UserRole{roles.USER_ADMIN}
		for i := range 5 {
			market := &domain.Market{
				UUID:         uuid.NewString(),
				Name:         fmt.Sprintf("tmarket_%d", i),
				Enabled:      false,
				AllowedRoles: roles.NewRoles(allowedRoles...),
				DeletedAt:    nil,
			}
			_, err := store.Save(context.Background(), market)

			assert.NoError(t, err)
		}

		roleInspector := guard.NewRoleInspector()
		spot := NewSpotInstrument(zap.NewNop(), store, roleInspector)
		result, _ := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Empty(t, result)
	})

	t.Run("should return markets allowed for user roles", func(t *testing.T) {
		store := ram.NewMarketStore()
		roleInspector := guard.NewRoleInspector()
		spot := NewSpotInstrument(zap.NewNop(), store, roleInspector)

		allowedRoles := []roles.UserRole{roles.USER_ADMIN, roles.USER_VERIFIED}

		market := &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "verif/admin",
			Enabled:      true,
			AllowedRoles: roles.NewRoles(allowedRoles...),
			DeletedAt:    nil,
		}
		_, err := store.Save(ctx, market)
		assert.NoError(t, err)

		allowedRoles2 := []roles.UserRole{roles.USER_ADMIN}

		market = &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "admin market",
			Enabled:      true,
			AllowedRoles: roles.NewRoles(allowedRoles2...),
			DeletedAt:    nil,
		}
		_, err = store.Save(ctx, market)

		assert.NoError(t, err)

		market = &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "deleted market",
			Enabled:      true,
			AllowedRoles: roles.NewRoles(allowedRoles...),
			DeletedAt:    nil,
		}
		_, err = store.Save(ctx, market)

		assert.NoError(t, err)
		spot.DeleteMarket(context.Background(), market.UUID)

		result, _ := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Len(t, result, 2)

		marketNames := []string{result[0].Name, result[1].Name}
		assert.Contains(t, marketNames, "verif/admin")
		assert.Contains(t, marketNames, "admin market")
	})

	t.Run("dont return disabled markets", func(t *testing.T) {
		store := ram.NewMarketStore()
		roleInspector := guard.NewRoleInspector()
		spot := NewSpotInstrument(zap.NewNop(), store, roleInspector)

		allowedRoles := []roles.UserRole{roles.USER_ADMIN}
		market := &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "enabled market",
			Enabled:      true,
			AllowedRoles: roles.NewRoles(allowedRoles...),
			DeletedAt:    nil,
		}
		_, err := store.Save(ctx, market)

		assert.NoError(t, err)

		market = &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "disabled market",
			Enabled:      false,
			AllowedRoles: roles.NewRoles(allowedRoles...),
			DeletedAt:    nil,
		}
		_, err = store.Save(ctx, market)

		assert.NoError(t, err)

		result, _ := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Len(t, result, 1)
		assert.Equal(t, "enabled market", result[0].Name)
	})
}

func TestSpotService_NewMarket(t *testing.T) {
	ctx := context.Background()

	t.Run("should create new market successfully", func(t *testing.T) {
		store := ram.NewMarketStore()
		roleInspector := guard.NewRoleInspector()
		spot := NewSpotInstrument(zap.NewNop(), store, roleInspector)

		allowedRoles := []roles.UserRole{roles.USER_ADMIN, roles.USER_VERIFIED}
		dto := &domain.CreateMarketDto{
			Name:         "new",
			Enabled:      true,
			AllowedRoles: allowedRoles,
		}
		result, err := spot.NewMarket(ctx, dto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "new", result.Name)
		assert.NotZero(t, result.UUID)

		markets, _ := spot.ViewMarkets(context.Background(), allowedRoles)

		require.Len(t, markets, 1)

		newMarket := markets[0]
		assert.Equal(t, result, newMarket)
	})
}

func TestSpotService_DeleteMarket(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete market successfully", func(t *testing.T) {
		store := ram.NewMarketStore()
		roleInspector := guard.NewRoleInspector()
		spotService := NewSpotInstrument(zap.NewNop(), store, roleInspector)

		newMarket := &domain.Market{
			UUID:         uuid.NewString(),
			Name:         "market_for_delete",
			Enabled:      true,
			AllowedRoles: roles.NewRoles(roles.USER_ADMIN),
		}

		market, err := store.Save(ctx, newMarket)
		assert.NoError(t, err)
		assert.NotZero(t, market.UUID)

		err = spotService.DeleteMarket(ctx, market.UUID)
		assert.NoError(t, err)

		_, err = store.Get(ctx, market.UUID)
		assert.Error(t, err)

		assert.True(t, market.IsDeleted())
		assert.False(t, market.IsEnabled())
	})

	t.Run("should return error when market not found", func(t *testing.T) {
		store := ram.NewMarketStore()
		roleInspector := guard.NewRoleInspector()
		spot := NewSpotInstrument(zap.NewNop(), store, roleInspector)

		uuid := "not-exist-uuid"
		err := spot.DeleteMarket(ctx, uuid)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNotFound)
	})
}
