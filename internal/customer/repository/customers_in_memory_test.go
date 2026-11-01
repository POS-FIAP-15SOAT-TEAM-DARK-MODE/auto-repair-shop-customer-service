package repository

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemory_SaveListDelete(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemory()

	c := &domain.Customer{
		ID: "c1", UserID: "u1", Type: domain.IndividualCustomerType, CPF: "52998224725", Phone: "1",
		User: &authDomain.User{ID: "u1", Name: "João", CoreUser: authDomain.CoreUser{Email: "j@s.com"}},
	}
	require.NoError(t, repo.Save(ctx, c))

	byID, err := repo.List(ctx, adapters.ListCustomerParams{ID: "c1"})
	require.NoError(t, err)
	require.Len(t, byID, 1)
	assert.Equal(t, "João", byID[0].User.Name)

	byDoc, err := repo.List(ctx, adapters.ListCustomerParams{Document: "52998224725"})
	require.NoError(t, err)
	assert.Len(t, byDoc, 1)

	byUser, err := repo.List(ctx, adapters.ListCustomerParams{UserID: "u1"})
	require.NoError(t, err)
	assert.Len(t, byUser, 1)

	none, err := repo.List(ctx, adapters.ListCustomerParams{Document: "00000000000"})
	require.NoError(t, err)
	assert.Empty(t, none)

	all, err := repo.List(ctx, adapters.ListCustomerParams{})
	require.NoError(t, err)
	assert.Len(t, all, 1)

	require.NoError(t, repo.Delete(ctx, "c1"))
	assert.ErrorIs(t, repo.Delete(ctx, "c1"), domain.ErrCustomerNotFound)
}

func TestInMemory_SaveConflictOnDocument(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemory()
	require.NoError(t, repo.Save(ctx, &domain.Customer{ID: "c1", CPF: "52998224725"}))

	assert.ErrorIs(t, repo.Save(ctx, &domain.Customer{ID: "c2", CPF: "52998224725"}), app.ErrDataConflict)
	require.NoError(t, repo.Save(ctx, &domain.Customer{ID: "c3", CNPJ: "x"}))
	require.NoError(t, repo.Save(ctx, &domain.Customer{ID: "c1", CPF: "52998224725", Phone: "2"}))
}
