package client

import (
	"main/api/spotpb"
	"main/order/domain"
	"main/pkg/roles"
)

type SpotClientMapper struct{}

// internal --- > protobuf
func (mapper *SpotClientMapper) ToPbRoles(roles []roles.UserRole) []spotpb.UserRole {
	out := make([]spotpb.UserRole, 0, len(roles))

	for _, r := range roles {
		out = append(out, spotpb.UserRole(r))
	}

	return out
}

// pb --- > internal
func (mapper *SpotClientMapper) FromPbToMarkets(pbmarkets []*spotpb.Market) []*domain.Market {
	out := make([]*domain.Market, 0, len(pbmarkets))

	for _, pbm := range pbmarkets {
		market := domain.NewMarket(pbm.Id, pbm.Name)
		out = append(out, market)
	}

	return out
}
