package model

import "time"

type Market struct {
	UUID         string     `json:"uuid"`
	Name         string     `json:"name"`
	IsEnabled    bool       `json:"is_enabled"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	AllowedRoles []UserRole `json:"allowed_roles"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (m *Market) IsActive() bool {
	return m.IsEnabled && m.DeletedAt == nil
}

func (m *Market) IsAccessibleForRoles(roles []UserRole) bool {
	if len(m.AllowedRoles) == 0 {
		return true
	}

	for _, ur := range roles {
		for _, ar := range m.AllowedRoles {
			if ur == ar {
				return true
			}
		}
	}
	return false
}
