package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapMarketsToProtoMarkets(markets []*model.Market) []*modelsv1.Market {
	out := make([]*modelsv1.Market, 0, len(markets))

	for _, m := range markets {
		out = append(out, MapMarketToProtoMarket(m))
	}

	return out
}

func MapMarketToProtoMarket(market *model.Market) *modelsv1.Market {
	return &modelsv1.Market{
		Uuid:      market.UUID,
		Name:      market.Name,
		IsActive:  market.IsEnabled,
		CreatedAt: timestamppb.New(market.CreatedAt),
		UpdatedAt: timestamppb.New(market.UpdatedAt),
	}
}
