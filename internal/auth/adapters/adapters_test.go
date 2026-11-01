package adapters_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/stretchr/testify/assert"
)

func TestChangeRoleRequest_IsValid(t *testing.T) {
	assert.NoError(t, adapters.ChangeRoleRequest{Role: domain.MECHANIC}.IsValid())
	assert.Error(t, adapters.ChangeRoleRequest{Role: "SUPERUSER"}.IsValid())
}

func TestLoginRequest_Domain(t *testing.T) {
	c := adapters.LoginRequest{Email: "a@b.co", Password: "Secret!1"}.Domain()
	assert.Equal(t, "a@b.co", c.Email)
	assert.Equal(t, "Secret!1", c.Password)
}
