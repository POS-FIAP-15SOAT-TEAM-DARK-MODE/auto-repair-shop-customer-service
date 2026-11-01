package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=AuthService --with-expecter
type AuthService interface {
	Login(ctx context.Context, data adapters.LoginRequest) (adapters.LoginResponse, error)
	Register(ctx context.Context, data adapters.CreateUserRequest) (adapters.UserResponse, error)
	ChangeRole(ctx context.Context, userId string, data adapters.ChangeRoleRequest) error
}
