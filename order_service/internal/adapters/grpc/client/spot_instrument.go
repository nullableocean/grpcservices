package client

import (
	"context"
	"errors"
	"time"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var _ ports.SpotInstrument = &SpotInstrumentClient{}

type SpotInstrumentClient struct {
	client spotv1.SpotInstrumentClient
	logger *logger.CtxZapLogger

	reqTimeout time.Duration
}

type Option struct {
	RequestTimeout time.Duration
}

func NewSpotInstrumentClient(l *logger.CtxZapLogger, grpcClient spotv1.SpotInstrumentClient, opts Option) (*SpotInstrumentClient, error) {
	if opts.RequestTimeout <= 0 {
		return nil, errors.New("invalid request timeout option")
	}

	return &SpotInstrumentClient{
		client:     grpcClient,
		logger:     l,
		reqTimeout: opts.RequestTimeout,
	}, nil
}

func (cl *SpotInstrumentClient) FindMarket(ctx context.Context, marketUuid string) (*model.Market, error) {
	ctx, span := otel.Tracer("spot_grpc_client").Start(ctx, "find_market")
	defer span.End()

	cl.logger.Debug(ctx, "call FindMarket from SpotInstrument grpc server")

	request := &spotv1.FindMarketRequest{
		MarketUuid: marketUuid,
	}

	res, err := cl.client.FindMarket(ctx, request)
	if err != nil {
		cl.logger.Error(ctx, "failed get market from spot instrument", zap.Error(err))

		return nil, mapping.MapGrpcStatusToError(err)
	}

	return mapping.MapProtoMarketToMarket(res.Market), nil
}
