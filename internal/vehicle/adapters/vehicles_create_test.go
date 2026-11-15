package adapters

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/stretchr/testify/assert"
)

func TestVehicleDomainToResponse(t *testing.T) {
	vehicle := domain.Vehicle{
		ID:           "vehicle-id",
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "customer-1",
	}

	got := VehicleDomainToResponse(vehicle)

	assert.Equal(t, VehicleResponse{
		ID:           "vehicle-id",
		LicensePlate: "ABC1D23",
		BrandModel:   "chevrolet - onix",
		Year:         2020,
		CustomerId:   "customer-1",
	}, got)
}
