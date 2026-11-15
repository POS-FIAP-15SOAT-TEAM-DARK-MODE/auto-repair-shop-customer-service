package adapters

import (
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

type CreateVehicle struct {
	LicensePlate string `json:"licensePlate"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	CustomerId   string `json:"customerId"`
}

type VehicleResponse struct {
	ID           string `json:"id"`
	LicensePlate string `json:"licensePlate"`
	BrandModel   string `json:"brandModel"`
	Year         int    `json:"year"`
	CustomerId   string `json:"customerId"`
}

func VehicleDomainToResponse(vehicle domain.Vehicle) VehicleResponse {
	return VehicleResponse{
		ID:           vehicle.ID,
		BrandModel:   fmt.Sprintf("%s - %s", vehicle.Brand, vehicle.Model),
		Year:         vehicle.Year,
		LicensePlate: vehicle.LicensePlate,
		CustomerId:   vehicle.CustomerId,
	}
}
