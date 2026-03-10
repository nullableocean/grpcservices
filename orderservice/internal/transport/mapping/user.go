package mapping

import (
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/shared/roles"
)

// Map proto user roles from response to service user roles
func MapProtoUserRolesResponseToUserRoles(pbrUserRolesResp *userv1.UserRolesResponse) []roles.UserRole {
	out := make([]roles.UserRole, 0, len(pbrUserRolesResp.UserRoles))
	for _, r := range pbrUserRolesResp.UserRoles {
		out = append(out, roles.UserRole(r))
	}

	return out
}

// UserRoles ---> Protobuff UserRoles
func MapUserRolesToProtoRoles(roles []roles.UserRole) []typesv1.UserRole {
	out := make([]typesv1.UserRole, 0, len(roles))

	for _, r := range roles {
		out = append(out, typesv1.UserRole(r))
	}

	return out
}
