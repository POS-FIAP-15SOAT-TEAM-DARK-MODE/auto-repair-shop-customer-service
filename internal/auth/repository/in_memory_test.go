package repository

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAuthRepo_Flow(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemory()

	u := &domain.User{ID: "u1", Name: "Admin", CoreUser: domain.CoreUser{Email: "a@b.co", Password: "x"}}
	require.NoError(t, repo.Save(ctx, u))
	assert.ErrorIs(t, repo.Save(ctx, &domain.User{ID: "u2", CoreUser: domain.CoreUser{Email: "a@b.co"}}), app.ErrDataConflict)

	require.NoError(t, repo.SaveUserRole(ctx, "u1", domain.ADMIN))

	found, err := repo.GetByEmail(ctx, "a@b.co")
	require.NoError(t, err)
	assert.Equal(t, "u1", found.ID)

	missing, err := repo.GetByEmail(ctx, "none@b.co")
	require.NoError(t, err)
	assert.Empty(t, missing.ID)

	session, err := repo.GetUserSession(ctx, "u1")
	require.NoError(t, err)
	assert.NotEmpty(t, session.Token)

	require.NoError(t, repo.Update(ctx, "u1", "Novo", "n@b.co"))
	assert.ErrorIs(t, repo.Update(ctx, "missing", "x", "y@b.co"), domain.ErrInvalidUserCredentials)

	require.NoError(t, repo.Delete(ctx, "u1"))
	assert.ErrorIs(t, repo.Delete(ctx, "u1"), domain.ErrInvalidUserCredentials)
}
