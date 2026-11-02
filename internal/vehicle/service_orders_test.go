package vehicle

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	uowMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	domainMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// noVehicleOrders answers that no vehicle has service orders.
type noVehicleOrders struct{}

func (noVehicleOrders) HasVehicleOrders(context.Context, string) (bool, error) { return false, nil }

// fakeVehicleOrders is a ServiceOrderChecker with a canned answer, recording the vehicle it was asked about.
type fakeVehicleOrders struct {
	has      bool
	err      error
	asked    string
	asCalled bool
}

func (f *fakeVehicleOrders) HasVehicleOrders(_ context.Context, vehicleID string) (bool, error) {
	f.asked = vehicleID
	f.asCalled = true
	return f.has, f.err
}

func TestService_Delete_ServiceOrderCheck(t *testing.T) {
	cases := []struct {
		name    string
		orders  *fakeVehicleOrders
		wantErr error
		deletes bool
	}{
		{"no service orders deletes", &fakeVehicleOrders{}, nil, true},
		{"service orders refuse with domain error", &fakeVehicleOrders{has: true}, domain.ErrVehicleHasServiceOrders, false},
		{"upstream unavailable does not delete", &fakeVehicleOrders{err: app.ErrUpstreamUnavailable}, app.ErrUpstreamUnavailable, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			exec := uowMocks.NewExecutor(t)
			repo := domainMocks.NewVehicleRepository(t)
			if tc.deletes {
				// Nothing is expected on repo or exec when the delete must not happen: a call would fail the test.
				exec.EXPECT().Execute(mock.Anything, mock.Anything).Return(nil)
			}
			service := NewService(exec, repo, id.NewNoopIDGenerator("mock-id"), tc.orders)

			err := service.Delete(ctx, "veh-1")
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, tc.orders.asCalled, "the Service Order must be consulted before deleting")
			assert.Equal(t, "veh-1", tc.orders.asked)
		})
	}
}
