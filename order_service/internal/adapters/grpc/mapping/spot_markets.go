package mapping

import (
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func MapProtoMarketsToMarkets(markets []*spotv1.Market) []*model.Market {
	out := make([]*model.Market, 0, len(markets))

	for _, m := range markets {
		out = append(out, MapProtoMarketToMarket(m))
	}

	return out
}

func MapProtoMarketToMarket(pbm *spotv1.Market) *model.Market {
	return &model.Market{
		UUID: pbm.Uuid,
	}
}
