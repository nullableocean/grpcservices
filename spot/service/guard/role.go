package guard

import (
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/domain"
)

type RoleInspector struct{}

func NewRoleInspector() *RoleInspector {
	return &RoleInspector{}
}

func (ri *RoleInspector) HasAccessToMarket(m *domain.Market, role roles.UserRole) bool {
	return m.IsAllowed(role)
}
