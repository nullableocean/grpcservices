package seed

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
	"go.uber.org/zap"
)

type SpotInstrument interface {
	NewMarket(ctx context.Context, dto *domain.CreateMarketDto) (*domain.Market, error)
}

func SeedMarkets(logger *zap.Logger, spot SpotInstrument) {
	rolesList := []roles.UserRole{
		roles.USER_GUEST,
		roles.USER_VERIFIED,
		roles.USER_SELLER,
		roles.USER_MODER,
		roles.USER_ADMIN,
	}

	rolesNames := []string{
		roles.MapInString(roles.USER_GUEST),
		roles.MapInString(roles.USER_VERIFIED),
		roles.MapInString(roles.USER_SELLER),
		roles.MapInString(roles.USER_MODER),
		roles.MapInString(roles.USER_ADMIN),
	}

	marketsName := []string{
		"BTC/ETH",
		"BTC/XRP",
		"BTC",
		"XRP",
		"ETH",
	}

	count := len(rolesList)
	ctx := context.Background()
	for i := range count {
		name := marketsName[i]

		dto := &domain.CreateMarketDto{
			Name:         name,
			Enabled:      true,
			AllowedRoles: rolesList[count-i-1:],
		}

		m, err := spot.NewMarket(ctx, dto)
		if err != nil {
			logger.Info("seed new market error", zap.Error(err))
			continue
		}

		logger.Info("seeded new market",
			zap.String("UUID", m.UUID),
			zap.String("name", m.Name),
			zap.Strings("allowed roles", rolesNames[count-i-1:]),
		)
	}
}
