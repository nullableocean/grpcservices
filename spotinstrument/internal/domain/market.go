package domain

import (
	"time"

	"github.com/nullableocean/grpcservices/shared/roles"
)

type Market struct {
	UUID    string
	Name    string
	Enabled bool

	AllowedRoles *roles.Roles
	DeletedAt    *time.Time
}

type CreateMarketDto struct {
	Name         string
	Enabled      bool
	AllowedRoles []roles.UserRole
}

func (m *Market) IsEnabled() bool {
	return m.Enabled
}

func (m *Market) Disable() {
	m.Enabled = false
}

func (m *Market) Enable() {
	m.Enabled = true
}

func (m *Market) IsAllowed(role roles.UserRole) bool {
	return m.AllowedRoles.Has(role)
}

func (m *Market) AddAllowedRole(role roles.UserRole) {
	m.AllowedRoles.Add(role)
}

func (m *Market) RemoveAllowedRole(role roles.UserRole) {
	m.AllowedRoles.Remove(role)
}

func (m *Market) IsDeleted() bool {
	return m.DeletedAt != nil
}

func (m *Market) GetDeletedAt() time.Time {
	return *m.DeletedAt
}

func (m *Market) Delete() {
	m.Disable()

	now := time.Now()
	m.DeletedAt = &now
}
