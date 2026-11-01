package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=AuthHTTPController --with-expecter
type AuthHTTPController interface {
	Login(ctx context.Context, req *http.Request) (adapters.LoginResponse, error)
	Register(ctx context.Context, req *http.Request) (adapters.UserResponse, error)
	ChangeRole(ctx context.Context, req *http.Request) error
}
