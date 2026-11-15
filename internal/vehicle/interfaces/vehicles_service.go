package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleService --with-expecter
type VehicleService interface {
	Create(ctx context.Context, vehicle adapters.CreateVehicle) (adapters.VehicleResponse, error)
	List(ctx context.Context, filter adapters.ListVehiclesParams) (adapters.PaginatedVehicleResponse, error)
	FindById(ctx context.Context, id string) (domain.Vehicle, error)
	Edit(ctx context.Context, id string, vehicle adapters.CreateVehicle) (adapters.VehicleResponse, error)
	Delete(ctx context.Context, id string) error
}
