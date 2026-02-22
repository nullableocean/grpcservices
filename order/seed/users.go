package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/nullableocean/grpcservices/order/service/user"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"go.uber.org/zap"
)

func SeedUsers(logger *zap.Logger, userService *user.UserService) {
	rolesList := []roles.UserRole{
		roles.USER_GUEST,
		roles.USER_VERIFIED,
		roles.USER_SELLER,
		roles.USER_MODER,
		roles.USER_ADMIN,
	}

	rolesNames := []string{
		roles.MapInString(roles.USER_GUEST),
		roles.MapInString(roles.USER_VERIFIED),
		roles.MapInString(roles.USER_SELLER),
		roles.MapInString(roles.USER_MODER),
		roles.MapInString(roles.USER_ADMIN),
	}

	count := len(rolesList)

	ctx := context.Background()
	for i := range count {
		username := fmt.Sprintf("user_%d", i)

		u, err := userService.CreateUser(ctx, username, genString(), rolesList[:count-i])
		if err != nil {
			logger.Info("seed new user error", zap.Error(err))
			continue
		}

		logger.Info("seeded new user success",
			zap.Int64("ID", u.Id()),
			zap.String("username", username),
			zap.Strings("roles", rolesNames[:count-i]),
		)
	}

}

func genString() string {
	buf := make([]byte, 8)
	rand.Read(buf)

	return base64.RawURLEncoding.EncodeToString(buf)
}
