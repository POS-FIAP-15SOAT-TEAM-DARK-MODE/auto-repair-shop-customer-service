package adapters

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
)

type ChangeRoleRequest struct {
	Role domain.Role `json:"role"`
}

func (c ChangeRoleRequest) IsValid() error {
	if _, ok := domain.ValidRoles[c.Role]; !ok {
		return domain.ErrInvalidUserCredentials
	}
	return nil
}

type CreateUserRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
