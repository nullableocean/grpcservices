package mapping

import (
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
)

func MapMarketsToProtoMarkets(markets []*model.Market) []*spotv1.Market {
	out := make([]*spotv1.Market, 0, len(markets))

	for _, m := range markets {
		out = append(out, MapMarketToProtoMarket(m))
	}

	return out
}

func MapMarketToProtoMarket(market *model.Market) *spotv1.Market {
	return &spotv1.Market{
		Uuid: market.UUID,
	}
}
