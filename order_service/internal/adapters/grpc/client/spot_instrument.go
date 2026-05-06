package client

import (
	"context"
	"time"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var _ ports.SpotInstrument = &SpotInstrumentClient{}

var (
	defaultTimeout = time.Second * 10
)

type SpotInstrumentClient struct {
	client spotv1.SpotInstrumentClient
	logger *zap.Logger

	reqTimeout time.Duration
}

type Option struct {
	RequestTimeout time.Duration
}

func NewSpotInstrumentClient(l *zap.Logger, grpcClient spotv1.SpotInstrumentClient, opts Option) *SpotInstrumentClient {
	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = defaultTimeout
	}

	return &SpotInstrumentClient{
		client:     grpcClient,
		logger:     l,
		reqTimeout: opts.RequestTimeout,
	}
}

func (cl *SpotInstrumentClient) ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]*model.Market, error) {
	ctx, span := otel.Tracer("spot_grpc_client").Start(ctx, "view_markets")
	defer span.End()

	cl.logger.Info("call ViewMarkets from SpotInstrument grpc server")

	request := &spotv1.ViewMarketsRequest{
		UserRoles: mapping.MapRolesToProtoUserRoles(userRoles),
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, cl.reqTimeout)
	defer cancel()

	response, err := cl.client.ViewMarkets(timeoutCtx, request)
	if err != nil {
		cl.logger.Error("failed get markets from grpc server", zap.Error(err))
		span.AddEvent("failed grpc call")

		return nil, mapping.MapGrpcStatusToError(err)
	}

	cl.logger.Info("got markets from grpc server")
	span.AddEvent("success grpc call")

	return mapping.MapProtoMarketsToMarkets(response.Markets), nil
}

func (cl *SpotInstrumentClient) FindMarket(ctx context.Context, marketUuid string, userRoles []model.UserRole) (*model.Market, error) {
	ctx, span := otel.Tracer("spot_grpc_client").Start(ctx, "find_markets")
	defer span.End()

	cl.logger.Info("call FindMarket from SpotInstrument grpc server")

	request := &spotv1.FindMarketRequest{
		MarketUuid: marketUuid,
		UserRoles:  mapping.MapRolesToProtoUserRoles(userRoles),
	}

	res, err := cl.client.FindMarket(ctx, request)
	if err != nil {
		cl.logger.Warn("failed get market from spot instrument", zap.Error(err))

		return nil, mapping.MapGrpcStatusToError(err)
	}

	return mapping.MapProtoMarketToMarket(res.Market), nil
}
