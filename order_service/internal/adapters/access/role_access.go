package access

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type RoleAccessService struct {
}

func NewRoleAccessService() *RoleAccessService {
	return &RoleAccessService{}
}

func (s *RoleAccessService) CanCreateOrder(ctx context.Context, user *model.User, params *dto.CreateOrderParameters) error {
	allowedRoles := map[model.UserRole]bool{
		model.UserRoleTrader:      true,
		model.UserRoleMarketMaker: true,
		model.UserRoleModer:       true,
		model.UserRoleAdmin:       true,
	}

	hasAllowedRole := false
	for _, r := range user.Roles {
		if allowedRoles[r] {
			hasAllowedRole = true
			break
		}
	}

	if !hasAllowedRole {
		return fmt.Errorf("user roles %v do not include any allowed role", user.Roles)
	}

	roleAllowedTypes := map[model.UserRole][]model.OrderType{
		model.UserRoleTrader:      {model.OrderTypeLimit, model.OrderTypeMarket},
		model.UserRoleMarketMaker: {model.OrderTypeLimit, model.OrderTypeMarket, model.OrderTypeStopLoss, model.OrderTypeTakeProfit},
		model.UserRoleAdmin:       {model.OrderTypeLimit, model.OrderTypeMarket, model.OrderTypeStopLoss, model.OrderTypeTakeProfit},
	}

	typeAllowed := false
	for _, r := range user.Roles {
		if allowedTypes, ok := roleAllowedTypes[r]; ok {
			for _, allowed := range allowedTypes {
				if allowed == params.Type {
					typeAllowed = true
					break
				}
			}
		}

		if typeAllowed {
			break
		}
	}

	if !typeAllowed {
		return fmt.Errorf("order type %s not allowed for roles %v", params.Type, user.Roles)
	}

	return nil
}
