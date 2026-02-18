package server

import (
	"github.com/nullableocean/grpcservices/api/spotpb"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/domain"
)

type SpotMapper struct {
}

// internal --- > protobuf
func (m *SpotMapper) ToPbMarkets(markets []*domain.Market) []*spotpb.Market {
	out := make([]*spotpb.Market, 0, len(markets))

	for _, market := range markets {
		out = append(out, m.ToPbMarket(market))
	}

	return out
}

func (m *SpotMapper) ToPbMarket(market *domain.Market) *spotpb.Market {
	return &spotpb.Market{
		Id:   market.Id(),
		Name: market.Name(),
	}
}

// protobuf --- > internal
func (s *SpotMapper) FromPbToRoles(pbRoles []spotpb.UserRole) []roles.UserRole {
	out := make([]roles.UserRole, 0, len(pbRoles))

	for _, pbr := range pbRoles {
		out = append(out, roles.UserRole(pbr))
	}

	return out
}
