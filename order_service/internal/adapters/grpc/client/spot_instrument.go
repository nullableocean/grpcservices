package client

import (
	"context"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// implements
// var _ ports.SpotInstrument = &SpotInstrumentGrpcClient{}

type SpotInstrumentClient struct {
	client spotv1.SpotInstrumentClient
	logger *zap.Logger
}

func NewSpotInstrumentClient(l *zap.Logger, grpcClient spotv1.SpotInstrumentClient) *SpotInstrumentClient {
	return &SpotInstrumentClient{
		client: grpcClient,
		logger: l,
	}
}

func (cl *SpotInstrumentClient) ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]model.Market, error) {
	ctx, span := otel.Tracer("spot_grpc_client").Start(ctx, "view_markets")
	defer span.End()

	cl.logger.Info("call ViewMarkets from SpotInstrument grpc server")

	request := &spotv1.ViewMarketsRequest{
		UserRoles: mapping.MapRolesToProtoUserRoles(userRoles),
	}

	response, err := cl.client.ViewMarkets(ctx, request)
	if err != nil {
		cl.logger.Error("failed get markets from grpc server", zap.Error(err))
		span.AddEvent("failed grpc call")

		return nil, mapping.MapGrpcStatusToError(err)
	}

	cl.logger.Info("got markets from grpc server")
	span.AddEvent("success grpc call")

	return mapping.MapProtoMarketsToMarkets(response.Markets), nil
}
