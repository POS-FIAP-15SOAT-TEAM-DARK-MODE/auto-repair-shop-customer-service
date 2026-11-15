package adapters

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

type ListVehiclesParams struct {
	VehicleID  string
	CustomerID string
	Plate      string
	PageSize   int64
	Page       int64
}

type PaginatedVehicleResponse app.PaginatedResponse[VehicleResponse]

func VehiclesDomainToResponse(vehicles []domain.Vehicle) []VehicleResponse {
	res := make([]VehicleResponse, len(vehicles))
	for i, vehicle := range vehicles {
		res[i] = VehicleDomainToResponse(vehicle)
	}
	return res
}
