package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/nullableocean/grpcservices/api/spotpb"
	"github.com/nullableocean/grpcservices/spot/service"
)

// check interface
var _ spotpb.SpotInstrumentServer = &SpotInstrumentServer{}

type SpotInstrumentServer struct {
	spotpb.UnimplementedSpotInstrumentServer

	service *service.SpotInstrument
	mapper  *SpotMapper

	logger *zap.Logger
}

func NewSpotInstrumentServer(logger *zap.Logger, service *service.SpotInstrument) *SpotInstrumentServer {
	return &SpotInstrumentServer{
		service: service,
		mapper:  &SpotMapper{},

		logger: logger,
	}
}

func (serv *SpotInstrumentServer) ViewMarkets(ctx context.Context, req *spotpb.ViewMarketsRequest) (*spotpb.ViewMarketsResponse, error) {
	userRoles := serv.mapper.FromPbToRoles(req.UserRoles)
	markets := serv.service.ViewMarkets(userRoles)

	resp := &spotpb.ViewMarketsResponse{
		Markets: serv.mapper.ToPbMarkets(markets),
	}

	return resp, nil
}
