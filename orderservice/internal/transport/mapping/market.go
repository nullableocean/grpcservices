package mapping

import (
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
)

// Map spotinstrument pb markets to domain market
func MapProtoMarketsToDomainMarkets(pbmarkets []*spotv1.Market) []*domain.Market {
	out := make([]*domain.Market, 0, len(pbmarkets))

	for _, pbm := range pbmarkets {
		market := &domain.Market{
			UUID: pbm.Uuid,
			Name: pbm.Name,
		}
		out = append(out, market)
	}

	return out
}

func MapDomainOrderToStockmarketProcessRequest(o *domain.Order) *stockmarketv1.ProcessOrderRequest {
	return &stockmarketv1.ProcessOrderRequest{
		Order: MapDomainOrderToProtoOrder(o),
	}
}
