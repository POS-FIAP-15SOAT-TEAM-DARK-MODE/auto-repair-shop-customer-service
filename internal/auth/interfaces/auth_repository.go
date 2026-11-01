package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=AuthRepository --with-expecter
type AuthRepository interface {
	Save(ctx context.Context, user *domain.User) error
	SaveUserRole(ctx context.Context, userID string, role domain.Role) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserSession(ctx context.Context, id string) (adapters.LoginResponse, error)
	Update(ctx context.Context, userID, name, email string) error
	Delete(ctx context.Context, userID string) error
}
