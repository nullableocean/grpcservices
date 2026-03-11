package spotinstrument

import (
	"context"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/shared/roles"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
)

type SpotClient struct {
	client spotv1.SpotInstrumentClient
	logger *zap.Logger
}

func NewSpotClient(logger *zap.Logger, client spotv1.SpotInstrumentClient) *SpotClient {
	return &SpotClient{
		client: client,
		logger: logger,
	}
}

func (cl *SpotClient) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	ctx, span := otel.Tracer("spotinstrument_client").Start(ctx, "view_markets_request")
	defer span.End()

	request := &spotv1.ViewMarketsRequest{
		UserRoles: mapping.MapUserRolesToProtoRoles(roles),
	}

	cl.logger.Info("request markets from spotinstrument grpc server")

	resp, err := cl.client.ViewMarkets(ctx, request)
	if err != nil {
		cl.logger.Error("failed request to spotinstrument server", zap.Error(err))

		return nil, err
	}

	markets := mapping.MapProtoMarketsToDomainMarkets(resp.Markets)
	return markets, nil
}
