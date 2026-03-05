package server

import (
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
)

type SpotMapper struct {
}

// internal --- > protobuf
func (m *SpotMapper) ToPbMarkets(markets []*domain.Market) []*spotv1.Market {
	out := make([]*spotv1.Market, 0, len(markets))

	for _, market := range markets {
		out = append(out, m.ToPbMarket(market))
	}

	return out
}

func (m *SpotMapper) ToPbMarket(market *domain.Market) *spotv1.Market {
	return &spotv1.Market{
		Uuid: market.UUID,
		Name: market.Name,
	}
}

// protobuf --- > internal
func (s *SpotMapper) FromPbToRoles(pbRoles []typesv1.UserRole) []roles.UserRole {
	out := make([]roles.UserRole, 0, len(pbRoles))

	for _, pbr := range pbRoles {
		out = append(out, roles.UserRole(pbr))
	}

	return out
}
