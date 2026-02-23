package auth

import (
	"slices"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/pkg/order"
	"github.com/nullableocean/grpcservices/pkg/roles"
)

type Permission string

const (
	Buy  Permission = "buy_perm"
	Sell Permission = "sell_perm"
)

type RoleAccessService struct {
	perms         map[Permission][]roles.UserRole
	orderTypePerm map[order.OrderType][]Permission
}

func NewRoleAccessService() *RoleAccessService {
	return &RoleAccessService{
		perms: map[Permission][]roles.UserRole{
			Buy:  {roles.USER_VERIFIED, roles.USER_SELLER, roles.USER_MODER, roles.USER_ADMIN},
			Sell: {roles.USER_SELLER, roles.USER_MODER, roles.USER_ADMIN},
		},
		orderTypePerm: map[order.OrderType][]Permission{
			order.ORDER_TYPE_BUY:  {Buy},
			order.ORDER_TYPE_SELL: {Sell},
		},
	}
}

func (ras *RoleAccessService) CanCreate(user *domain.User, orderType order.OrderType) bool {
	permissions := ras.orderTypePerm[orderType]

	for _, p := range permissions {
		ok := slices.ContainsFunc(ras.perms[p], user.HasRole)

		if !ok {
			return false
		}
	}

	return true
}
