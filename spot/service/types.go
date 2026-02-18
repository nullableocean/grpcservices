package service

import (
	"time"

	"github.com/nullableocean/grpcservices/pkg/roles"
)

type Market struct {
	id           int64
	name         string
	enabled      bool
	deletedAt    *time.Time
	allowedRoles map[roles.UserRole]struct{}
}

func (m *Market) Id() int64 {
	return m.id
}

func (m *Market) Name() string {
	return m.name
}

func (m *Market) IsEnabled() bool {
	return m.enabled
}

func (m *Market) IsAllowed(role roles.UserRole) bool {
	_, ex := m.allowedRoles[role]
	return ex
}

func (m *Market) IsDeleted() bool {
	return m.deletedAt != nil
}
