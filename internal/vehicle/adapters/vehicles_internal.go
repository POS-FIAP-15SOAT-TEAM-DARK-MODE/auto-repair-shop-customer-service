package adapters

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"

// InternalVehicleResponse is the payload of GET /internal/vehicles/{id}, consumed by service-order.
type InternalVehicleResponse struct {
	ID           string `json:"id"`
	CustomerID   string `json:"customer_id"`
	LicensePlate string `json:"license_plate"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
}

func VehicleDomainToInternalResponse(v domain.Vehicle) InternalVehicleResponse {
	return InternalVehicleResponse{
		ID:           v.ID,
		CustomerID:   v.CustomerId,
		LicensePlate: v.LicensePlate,
		Brand:        v.Brand,
		Model:        v.Model,
		Year:         v.Year,
	}
}
