package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func MapProtoMarketsToMarkets(markets []*modelsv1.Market) []*model.Market {
	out := make([]*model.Market, 0, len(markets))

	for _, m := range markets {
		out = append(out, MapProtoMarketToMarket(m))
	}

	return out
}

func MapProtoMarketToMarket(pbm *modelsv1.Market) *model.Market {
	return &model.Market{
		UUID:      pbm.Uuid,
		Name:      pbm.Name,
		IsActive:  pbm.IsActive,
		CreatedAt: pbm.CreatedAt.AsTime(),
		UpdatedAt: pbm.UpdatedAt.AsTime(),
	}
}
