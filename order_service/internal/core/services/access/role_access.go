package access

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
)

var _ ports.AccessService = &AccessService{}

type AccessService struct {
}

func NewRoleAccessService() *AccessService {
	return &AccessService{}
}

func (s *AccessService) CanCreateOrder(ctx context.Context, user *model.User, params *dto.CreateOrderParameters) error {
	if err := s.checkAllowedRoles(user); err != nil {
		return err
	}

	if err := s.checkAllowedTypeForRoles(user, params.Type); err != nil {
		return err
	}

	return nil
}

func (s *AccessService) checkAllowedRoles(user *model.User) error {
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

	return nil
}

func (s *AccessService) checkAllowedTypeForRoles(user *model.User, t model.OrderType) error {

	roleAllowedTypes := map[model.UserRole][]model.OrderType{
		model.UserRoleTrader:      {model.OrderTypeLimit, model.OrderTypeMarket},
		model.UserRoleMarketMaker: {model.OrderTypeLimit, model.OrderTypeMarket, model.OrderTypeStopLoss, model.OrderTypeTakeProfit},
		model.UserRoleAdmin:       {model.OrderTypeLimit, model.OrderTypeMarket, model.OrderTypeStopLoss, model.OrderTypeTakeProfit},
	}

	typeAllowed := false
	for _, r := range user.Roles {
		if allowedTypes, ok := roleAllowedTypes[r]; ok {
			for _, allowed := range allowedTypes {
				if allowed == t {
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
		return fmt.Errorf("order type %s not allowed for roles %v", t, user.Roles)
	}

	return nil
}
