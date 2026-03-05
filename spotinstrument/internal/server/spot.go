package server

import (
	"context"
	"time"

	"go.uber.org/zap"

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
	userRoles := serv.mapper.FromPbToRoles(req.UserRoles)

	serv.logger.Info("view market request", zap.Strings("roles", roles.MapSliceToStrings(userRoles)))

	start := time.Now()
	defer func() {
		serv.metrics.CalledViewMarket(time.Since(start))
	}()

	markets := serv.service.ViewMarkets(ctx, userRoles)
	resp := &spotv1.ViewMarketsResponse{
		Markets: serv.mapper.ToPbMarkets(markets),
	}

	return resp, nil
}
