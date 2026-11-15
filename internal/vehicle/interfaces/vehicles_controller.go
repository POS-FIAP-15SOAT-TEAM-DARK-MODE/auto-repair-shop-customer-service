package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleHTTPController --with-expecter
type VehicleHTTPController interface {
	Create(ctx context.Context, req *http.Request) (adapters.VehicleResponse, error)
	List(ctx context.Context, req *http.Request) (adapters.PaginatedVehicleResponse, error)
	Edit(ctx context.Context, req *http.Request) (adapters.VehicleResponse, error)
	Delete(ctx context.Context, req *http.Request) error
}

// VehicleInternalHTTPController serves the service-to-service routes under /internal.
type VehicleInternalHTTPController interface {
	GetInternalByID(ctx context.Context, req *http.Request) (adapters.InternalVehicleResponse, error)
}
