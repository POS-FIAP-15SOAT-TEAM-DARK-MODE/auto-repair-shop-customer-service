package vehicle

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetInternalByID_ReturnsVehicle(t *testing.T) {
	svc := mocks.NewVehicleService(t)
	svc.EXPECT().FindById(mock.Anything, "v1").Return(domain.Vehicle{
		ID: "v1", LicensePlate: "ABC1D23", Brand: "Fiat", Model: "Uno", Year: 2018, CustomerId: "c1",
	}, nil)

	ctx := context.WithValue(context.Background(), idParam, "v1") //nolint:staticcheck
	got, err := NewController(svc).GetInternalByID(ctx, &http.Request{})
	require.NoError(t, err)
	assert.Equal(t, adapters.InternalVehicleResponse{
		ID: "v1", CustomerID: "c1", LicensePlate: "ABC1D23", Brand: "Fiat", Model: "Uno", Year: 2018,
	}, got)
}

func TestGetInternalByID_NotFound(t *testing.T) {
	svc := mocks.NewVehicleService(t)
	svc.EXPECT().FindById(mock.Anything, "missing").Return(domain.Vehicle{}, domain.ErrVehicleNotFound)

	ctx := context.WithValue(context.Background(), idParam, "missing") //nolint:staticcheck
	_, err := NewController(svc).GetInternalByID(ctx, &http.Request{})
	assert.True(t, errors.Is(err, domain.ErrVehicleNotFound))
}
