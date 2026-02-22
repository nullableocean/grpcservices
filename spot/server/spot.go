package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nullableocean/grpcservices/api/spotpb"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/service/metrics"
	"github.com/nullableocean/grpcservices/spot/service/spot"
)

// check interface
var _ spotpb.SpotInstrumentServer = &SpotInstrumentServer{}

type SpotInstrumentServer struct {
	spotpb.UnimplementedSpotInstrumentServer

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

func (serv *SpotInstrumentServer) ViewMarkets(ctx context.Context, req *spotpb.ViewMarketsRequest) (*spotpb.ViewMarketsResponse, error) {
	userRoles := serv.mapper.FromPbToRoles(req.UserRoles)

	serv.logger.Info("view market request", zap.Strings("roles", roles.MapSliceToStrings(userRoles)))
	start := time.Now()
	defer func() {
		serv.metrics.CalledViewMarket(time.Since(start))
	}()

	markets := serv.service.ViewMarkets(ctx, userRoles)
	resp := &spotpb.ViewMarketsResponse{
		Markets: serv.mapper.ToPbMarkets(markets),
	}

	return resp, nil
}
