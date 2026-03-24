package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func MapRolesToProtoUserRoles(userRoles []model.UserRole) []modelsv1.UserRole {
	out := make([]modelsv1.UserRole, 0, len(userRoles))

	for _, r := range userRoles {
		var role modelsv1.UserRole
		switch r {
		case model.UserRoleGuest:
			role = modelsv1.UserRole_USER_ROLE_GUEST
		case model.UserRoleTrader:
			role = modelsv1.UserRole_USER_ROLE_TRADER
		case model.UserRoleMarketMaker:
			role = modelsv1.UserRole_USER_ROLE_MARKET_MAKER
		case model.UserRoleModer:
			role = modelsv1.UserRole_USER_ROLE_MODER
		case model.UserRoleAdmin:
			role = modelsv1.UserRole_USER_ROLE_ADMIN
		}

		out = append(out, role)
	}

	return out
}

func MapProtoUserRolesToRoles(pbRoles []modelsv1.UserRole) []model.UserRole {
	rlsList := make([]model.UserRole, 0, len(pbRoles))

	for _, pbr := range pbRoles {
		var role model.UserRole
		switch pbr {
		case modelsv1.UserRole_USER_ROLE_GUEST:
			role = model.UserRoleGuest
		case modelsv1.UserRole_USER_ROLE_TRADER:
			role = model.UserRoleTrader
		case modelsv1.UserRole_USER_ROLE_MARKET_MAKER:
			role = model.UserRoleMarketMaker
		case modelsv1.UserRole_USER_ROLE_MODER:
			role = model.UserRoleModer
		case modelsv1.UserRole_USER_ROLE_ADMIN:
			role = model.UserRoleAdmin
		}

		rlsList = append(rlsList, role)
	}

	return rlsList
}
