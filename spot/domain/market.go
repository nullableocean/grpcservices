package domain

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/nullableocean/grpcservices/pkg/roles"
)

type Market struct {
	id      int64
	name    string
	enabled *atomic.Bool

	allowedRoles map[roles.UserRole]struct{}
	mu           sync.RWMutex

	deletedAt time.Time
	deleted   int32
}

type CreateMarketDto struct {
	Name         string
	Enabled      bool
	AllowedRoles map[roles.UserRole]struct{}
}

func NewMarket(id int64, dto *CreateMarketDto) *Market {
	enabled := &atomic.Bool{}
	enabled.Store(dto.Enabled)

	return &Market{
		id:           id,
		name:         dto.Name,
		enabled:      enabled,
		allowedRoles: dto.AllowedRoles,
		mu:           sync.RWMutex{},
		deletedAt:    time.Time{},
		deleted:      0,
	}
}

func (m *Market) Id() int64 {
	return m.id
}

func (m *Market) Name() string {
	return m.name
}

func (m *Market) IsEnabled() bool {
	return m.enabled.Load()
}

func (m *Market) Disable() {
	m.enabled.Store(false)
}

func (m *Market) Enable() {
	m.enabled.Store(true)
}

func (m *Market) IsAllowed(role roles.UserRole) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ex := m.allowedRoles[role]
	return ex
}

func (m *Market) AddAllowedRole(role roles.UserRole) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.allowedRoles == nil {
		m.allowedRoles = make(map[roles.UserRole]struct{})
	}

	m.allowedRoles[role] = struct{}{}
}

func (m *Market) RemoveAllowedRole(role roles.UserRole) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.allowedRoles == nil {
		m.allowedRoles = make(map[roles.UserRole]struct{})
		return
	}

	delete(m.allowedRoles, role)
}

func (m *Market) IsDeleted() bool {
	return atomic.LoadInt32(&m.deleted) == 1
}

func (m *Market) DeletedAt() time.Time {
	return m.deletedAt
}

func (m *Market) Delete() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if atomic.CompareAndSwapInt32(&m.deleted, 0, 1) {
		m.deletedAt = time.Now()
		m.enabled.Store(false)
	}
}
