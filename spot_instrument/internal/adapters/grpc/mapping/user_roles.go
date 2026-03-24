package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
)

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
