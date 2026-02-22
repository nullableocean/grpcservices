package spot

import (
	"context"
	"fmt"
	"testing"

	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/domain"
	"github.com/nullableocean/grpcservices/spot/service"
	"github.com/nullableocean/grpcservices/spot/service/store/ram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpotService_ViewMarkets(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty slice when no markets enabled", func(t *testing.T) {
		store := ram.NewMarketStore()

		allowedRoles := map[roles.UserRole]struct{}{
			roles.USER_ADMIN: {},
		}
		for i := range 5 {
			_, err := store.Save(context.Background(), &domain.CreateMarketDto{
				Name:         fmt.Sprintf("tmarket_%d", i),
				Enabled:      false,
				AllowedRoles: allowedRoles,
			})

			assert.NoError(t, err)
		}

		spot := NewSpotInstrument(store)
		result := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Empty(t, result)
	})

	t.Run("should return markets allowed for user roles", func(t *testing.T) {
		store := ram.NewMarketStore()
		spot := NewSpotInstrument(store)

		allowedRoles := map[roles.UserRole]struct{}{
			roles.USER_ADMIN:    {},
			roles.USER_VERIFIED: {},
		}
		_, err := store.Save(ctx, &domain.CreateMarketDto{
			Name:         "verif/admin",
			Enabled:      true,
			AllowedRoles: allowedRoles,
		})
		assert.NoError(t, err)

		allowedRoles2 := map[roles.UserRole]struct{}{
			roles.USER_ADMIN: {},
		}
		_, err = store.Save(ctx, &domain.CreateMarketDto{
			Name:         "admin market",
			Enabled:      true,
			AllowedRoles: allowedRoles2,
		})
		assert.NoError(t, err)

		deletedMarket, err := store.Save(ctx, &domain.CreateMarketDto{
			Name:         "deleted market",
			Enabled:      true,
			AllowedRoles: allowedRoles,
		})
		assert.NoError(t, err)
		deletedMarket.Delete()

		result := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Len(t, result, 2)

		marketNames := []string{result[0].Name(), result[1].Name()}
		assert.Contains(t, marketNames, "verif/admin")
		assert.Contains(t, marketNames, "admin market")
	})

	t.Run("dont return disabled markets", func(t *testing.T) {
		store := ram.NewMarketStore()
		spot := NewSpotInstrument(store)

		allowedRoles := map[roles.UserRole]struct{}{
			roles.USER_ADMIN: {},
		}

		_, err := store.Save(ctx, &domain.CreateMarketDto{
			Name:         "enabled market",
			Enabled:      true,
			AllowedRoles: allowedRoles,
		})
		assert.NoError(t, err)

		_, err = store.Save(ctx, &domain.CreateMarketDto{
			Name:         "disabled market",
			Enabled:      false,
			AllowedRoles: allowedRoles,
		})
		assert.NoError(t, err)

		result := spot.ViewMarkets(ctx, []roles.UserRole{roles.USER_ADMIN})
		assert.Len(t, result, 1)
		assert.Equal(t, "enabled market", result[0].Name())
	})
}

func TestSpotService_NewMarket(t *testing.T) {
	ctx := context.Background()

	t.Run("should create new market successfully", func(t *testing.T) {
		store := ram.NewMarketStore()
		spot := NewSpotInstrument(store)

		allowedRoles := []roles.UserRole{roles.USER_ADMIN, roles.USER_VERIFIED}
		result, err := spot.NewMarket(ctx, "new", allowedRoles)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "new", result.Name())
		assert.NotZero(t, result.Id())

		markets := spot.ViewMarkets(context.Background(), allowedRoles)

		require.Len(t, markets, 1)

		newMarket := markets[0]
		assert.Equal(t, result, newMarket)
	})
}

func TestSpotService_DeleteMarket(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete market successfully", func(t *testing.T) {
		store := ram.NewMarketStore()
		service := NewSpotInstrument(store)

		market, err := store.Save(ctx, &domain.CreateMarketDto{
			Name:    "market_for_delete",
			Enabled: true,
			AllowedRoles: map[roles.UserRole]struct{}{
				roles.USER_ADMIN: {},
			},
		})
		assert.NoError(t, err)
		assert.NotZero(t, market.Id())

		err = service.DeleteMarket(ctx, market.Id())
		assert.NoError(t, err)

		_, err = store.GetById(ctx, market.Id())
		assert.Error(t, err)

		assert.True(t, market.IsDeleted())
		assert.False(t, market.IsEnabled())
	})

	t.Run("should return error when market not found", func(t *testing.T) {
		store := ram.NewMarketStore()
		spot := NewSpotInstrument(store)

		marketId := int64(999)
		err := spot.DeleteMarket(ctx, marketId)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNotFound)
	})
}
