package server

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/metrics"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/spot"
)

type SpotInstrumentServer struct {
	spotv1.UnimplementedSpotInstrumentServer

	service *spot.SpotInstrument
	mapper  *SpotMapper

	metrics *metrics.SpotMetrics
	logger  *zap.Logger
}

func NewSpotInstrumentServer(service *spot.SpotInstrument, logger *zap.Logger, metrics *metrics.SpotMetrics) *SpotInstrumentServer {
	return &SpotInstrumentServer{
		service: service,
		mapper:  &SpotMapper{},

		metrics: metrics,
		logger:  logger,
	}
}

func (serv *SpotInstrumentServer) ViewMarkets(ctx context.Context, req *spotv1.ViewMarketsRequest) (*spotv1.ViewMarketsResponse, error) {
	ctx, span := otel.Tracer("spotinstrument_server").Start(ctx, "view_markets")
	defer span.End()

	userRoles := serv.mapper.FromPbToRoles(req.UserRoles)

	serv.logger.Info("view market request", zap.Strings("roles", roles.MapSliceToStrings(userRoles)))

	start := time.Now()
	defer func() {
		serv.metrics.CalledViewMarket(time.Since(start))
	}()

	markets, err := serv.service.ViewMarkets(ctx, userRoles)
	if err != nil {
		return nil, serv.getGrpcError(err)
	}

	resp := &spotv1.ViewMarketsResponse{
		Markets: serv.mapper.ToPbMarkets(markets),
	}

	return resp, nil
}

func (serv *SpotInstrumentServer) getGrpcError(e error) error {
	return status.Error(codes.Internal, e.Error())
}
