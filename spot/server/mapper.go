package server

import (
	"main/api/spotpb"
	pkg "main/pkg/roles"
	"main/spot/service"
)

type SpotMapper struct {
}

// internal --- > protobuf
func (m *SpotMapper) ToPbMarkets(markets []*service.Market) []*spotpb.Market {
	out := make([]*spotpb.Market, 0, len(markets))

	for _, market := range markets {
		out = append(out, m.ToPbMarket(market))
	}

	return out
}

func (m *SpotMapper) ToPbMarket(market *service.Market) *spotpb.Market {
	return &spotpb.Market{
		Id:   market.Id(),
		Name: market.Name(),
	}
}

// protobuf --- > internal
func (s *SpotMapper) FromPbToRoles(pbRoles []spotpb.UserRole) []pkg.UserRole {
	out := make([]pkg.UserRole, 0, len(pbRoles))

	for _, pbr := range pbRoles {
		out = append(out, pkg.UserRole(pbr))
	}

	return out
}
