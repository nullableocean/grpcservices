package server

import (
	"context"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SpotInstrumentServer struct {
	spotv1.UnimplementedSpotInstrumentServer

	spotInstrument *spotinstrument.SpotInstrument
	logger         *zap.Logger
}

func NewSpotInstrumentServer(l *zap.Logger, spotInstrument *spotinstrument.SpotInstrument) *SpotInstrumentServer {
	return &SpotInstrumentServer{
		spotInstrument: spotInstrument,
		logger:         l,
	}
}

func (srv *SpotInstrumentServer) FindMarket(ctx context.Context, req *spotv1.FindMarketRequest) (*spotv1.FindMarketResponse, error) {
	ctx, span := otel.Tracer("spot_instrument_server").Start(ctx, "find_market")
	defer span.End()

	userUUID, ok := shared_inters.UserUUIDFromContext(ctx)
	if !ok || userUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "user uuid not provided")
	}

	logger := srv.logger.With(zap.String("user_uuid", userUUID))

	logger.Debug("got grpc call FindMarket in SpotInstrumentServer")

	roles := mapping.MapProtoUserRolesToRoles(req.UserRoles)

	market, err := srv.spotInstrument.FindWithRoles(ctx, req.MarketUuid, roles)
	if err != nil {
		span.AddEvent("failed find market")
		logger.Error("failed find market", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	span.AddEvent("success find markets")
	logger.Debug("market found", zap.String("market_uuid", market.UUID))

	response := &spotv1.FindMarketResponse{
		Market: mapping.MapMarketToProtoMarket(market),
	}

	return response, nil
}

func (srv *SpotInstrumentServer) ViewMarkets(ctx context.Context, req *spotv1.ViewMarketsRequest) (*spotv1.ViewMarketsResponse, error) {
	ctx, span := otel.Tracer("spot_instrument_server").Start(ctx, "view_markets")
	defer span.End()

	user, err := srv.extractUserFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not extracted from context")
	}

	logger := srv.logger.With(zap.String("user_uuid", user.UUID))

	logger.Debug("got grpc call ViewMarkets in SpotInstrumentServer")

	pageToken := model.PageToken{
		Token: req.PageToken,
	}

	data, err := srv.spotInstrument.ViewMarketsPaginated(ctx, user, pageToken, req.PageSize)
	if err != nil {
		span.AddEvent("failed view markets")
		logger.Error("failed get markets", zap.Error(err))

		return nil, srv.getGrpcError(err)
	}

	span.AddEvent("success find markets")
	logger.Debug("response markets", zap.Int("markets_count", len(data.Markets)))

	return srv.mapMarketsToResponse(data.Markets, data.NextPageToken.Token), nil
}

func (srv *SpotInstrumentServer) extractUserFromCtx(ctx context.Context) (*model.User, error) {
	userUUID, ok := shared_inters.UserUUIDFromContext(ctx)
	if !ok || userUUID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	ctxRoles, ok := shared_inters.RolesFromContext(ctx)
	if !ok {
		srv.logger.Warn("roles not provided in context")
	}

	var roles []model.UserRole
	if len(ctxRoles) > 0 {
		roles = make([]model.UserRole, len(ctxRoles))
		for i, roleStr := range ctxRoles {
			roles[i] = model.UserRole(roleStr)
		}
	}

	user := model.NewUser(userUUID, roles)
	return user, nil
}

func (srv *SpotInstrumentServer) mapMarketsToResponse(markets []*model.Market, nextPageToken string) *spotv1.ViewMarketsResponse {
	return &spotv1.ViewMarketsResponse{
		Markets:       mapping.MapMarketsToProtoMarkets(markets),
		NextPageToken: nextPageToken,
	}
}

func (srv *SpotInstrumentServer) getGrpcError(e error) error {
	return mapping.MapErrorToGrpcStatusError(e)
}
