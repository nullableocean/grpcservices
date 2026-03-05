package transport

import (
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/userservice/internal/domain"
)

func MapProtoCreateRequestToCreateDto(pbreq *userv1.CreateUserRequest) *domain.CreateUserDto {
	return &domain.CreateUserDto{
		Username: pbreq.Username,
		Password: pbreq.Password,
	}
}

func MapUserToCreateResponse(user *domain.User) *userv1.CreateUserResponse {
	return &userv1.CreateUserResponse{
		UserUuid:  user.UUID,
		UserRoles: MapUserRolesToProtoRoles(user.Roles.GetSlice()),
	}
}

func MapUserRolesToProtoRoles(roles []roles.UserRole) []typesv1.UserRole {
	out := make([]typesv1.UserRole, 0, len(roles))

	for _, r := range roles {
		out = append(out, typesv1.UserRole(r))
	}

	return out
}
