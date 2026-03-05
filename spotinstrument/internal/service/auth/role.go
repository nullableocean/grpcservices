package guard

import (
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
)

type RoleInspector struct{}

func NewRoleInspector() *RoleInspector {
	return &RoleInspector{}
}

func (ri *RoleInspector) HasAccessToMarket(m *domain.Market, role roles.UserRole) bool {
	return m.IsAllowed(role)
}
