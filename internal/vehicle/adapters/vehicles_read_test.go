package adapters

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/stretchr/testify/assert"
)

func TestVehiclesDomainToResponse(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		got := VehiclesDomainToResponse(nil)
		assert.Empty(t, got)
	})

	t.Run("maps all vehicles preserving order", func(t *testing.T) {
		vehicles := []domain.Vehicle{
			{
				ID:           "id-1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			{
				ID:           "id-2",
				LicensePlate: "XYZ9A87",
				Brand:        "fiat",
				Model:        "uno",
				Year:         2015,
				CustomerId:   "customer-2",
			},
		}

		got := VehiclesDomainToResponse(vehicles)

		assert.Equal(t, []VehicleResponse{
			{
				ID:           "id-1",
				LicensePlate: "ABC1D23",
				BrandModel:   "chevrolet - onix",
				Year:         2020,
				CustomerId:   "customer-1",
			},
			{
				ID:           "id-2",
				LicensePlate: "XYZ9A87",
				BrandModel:   "fiat - uno",
				Year:         2015,
				CustomerId:   "customer-2",
			},
		}, got)
	})
}
