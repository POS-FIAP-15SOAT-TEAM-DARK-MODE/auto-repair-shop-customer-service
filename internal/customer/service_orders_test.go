package customer

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOrders is a ServiceOrderChecker with a canned answer, recording the customer it was asked about.
type fakeOrders struct {
	has      bool
	err      error
	asked    string
	asCalled bool
}

func (f *fakeOrders) HasCustomerOrders(_ context.Context, customerID string) (bool, error) {
	f.asked = customerID
	f.asCalled = true
	return f.has, f.err
}

func TestService_Delete_ServiceOrderCheck(t *testing.T) {
	cases := []struct {
		name        string
		orders      *fakeOrders
		wantErr     error
		wantDeleted bool
	}{
		{"no service orders deletes", &fakeOrders{has: false}, nil, true},
		{"service orders refuse with domain error", &fakeOrders{has: true}, domain.ErrCustomerHasServiceOrders, false},
		{"upstream unavailable does not delete", &fakeOrders{err: app.ErrUpstreamUnavailable}, app.ErrUpstreamUnavailable, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			userRepo := authRepo.NewInMemory()
			svc := NewService(uow.NewInMemoryUoW(), customerRepo.NewInMemory(), userRepo, tc.orders).(*customerService)

			created, err := svc.Create(ctx, adapters.CreateCustomerRequest{
				Name: "João Silva", Email: "joao@email.com", Type: "INDIVIDUAL",
				Document: "529.982.247-25", Phone: "11999990001",
			})
			require.NoError(t, err)

			err = svc.Delete(ctx, created.ID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, tc.orders.asCalled, "the Service Order must be consulted before deleting")
			assert.Equal(t, created.ID, tc.orders.asked)

			_, getErr := svc.GetByID(ctx, created.ID)
			if tc.wantDeleted {
				assert.ErrorIs(t, getErr, domain.ErrCustomerNotFound)
			} else {
				assert.NoError(t, getErr, "the customer must still exist")
			}
		})
	}
}

func TestService_Delete_MissingCustomerSkipsCheck(t *testing.T) {
	orders := &fakeOrders{}
	svc := NewService(uow.NewInMemoryUoW(), customerRepo.NewInMemory(), authRepo.NewInMemory(), orders)

	assert.NoError(t, svc.Delete(context.Background(), "00000000-0000-0000-0000-000000000001"))
	assert.False(t, orders.asCalled)
}
