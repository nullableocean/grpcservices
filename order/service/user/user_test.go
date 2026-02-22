package user_test

import (
	"context"
	"testing"

	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/store/ram"
	"github.com/nullableocean/grpcservices/order/service/user"
	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	store := ram.NewUserStore()
	usService := user.NewUserService(store)

	username := "testuser"
	password := "testpassword123"
	userRoles := []roles.UserRole{roles.USER_VERIFIED}

	createdUser, err := usService.CreateUser(context.Background(), username, password, userRoles)

	assert.NoError(t, err)
	assert.Equal(t, createdUser.Username(), username)
}

func TestUserService_GetUser(t *testing.T) {
	store := ram.NewUserStore()
	usService := user.NewUserService(store)

	username := "testuser"
	password := "testpassword123"
	userRoles := []roles.UserRole{roles.USER_VERIFIED}

	createdUser, err := usService.CreateUser(context.Background(), username, password, userRoles)

	assert.NoError(t, err)
	assert.Equal(t, createdUser.Username(), username)

	gotUser, err := usService.GetUser(context.Background(), createdUser.Id())
	assert.NoError(t, err)
	assert.Equal(t, createdUser, gotUser)
}

func TestUserService_DeleteUser(t *testing.T) {
	store := ram.NewUserStore()
	usService := user.NewUserService(store)

	username := "testuser"
	password := "testpassword123"
	userRoles := []roles.UserRole{roles.USER_VERIFIED}

	createdUser, err := usService.CreateUser(context.Background(), username, password, userRoles)

	assert.NoError(t, err)
	assert.Equal(t, createdUser.Username(), username)

	gotUser, err := usService.GetUser(context.Background(), createdUser.Id())
	assert.NoError(t, err)
	assert.Equal(t, createdUser, gotUser)

	err = usService.DeleteUser(context.Background(), createdUser.Id())
	assert.NoError(t, err)
	assert.True(t, createdUser.IsDeleted())

	err = usService.DeleteUser(context.Background(), createdUser.Id())
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNotFound)

	gotUser, err = usService.GetUser(context.Background(), createdUser.Id())
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNotFound)
}
