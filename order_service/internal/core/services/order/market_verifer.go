package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

type MarketValidator struct {
	spotInstrument ports.SpotInstrument
	logger         *zap.Logger
}

func NewMarketValidator(spotInstrument ports.SpotInstrument, logger *zap.Logger) *MarketValidator {
	return &MarketValidator{
		spotInstrument: spotInstrument,
		logger:         logger,
	}
}

func (v *MarketValidator) Validate(ctx context.Context, marketUUID string, user *model.User) error {
	market, err := v.spotInstrument.FindMarket(ctx, marketUUID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return fmt.Errorf("market not found: %w", errs.ErrNotFound)
		}
		if errors.Is(err, ports.ErrNotAllowed) {
			return fmt.Errorf("market not allowed for user roles: %w", errs.ErrNotAllowed)
		}

		return fmt.Errorf("failed to validate market: %w", err)
	}

	if !market.IsActive {
		return fmt.Errorf("market is not active: %w", errs.ErrNotAllowed)
	}

	return nil
}
