package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleRepository --with-expecter
type VehicleRepository interface {
	Save(ctx context.Context, vehicle *domain.Vehicle) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, params adapters.ListVehiclesParams) ([]domain.Vehicle, error)
	Count(ctx context.Context, params adapters.ListVehiclesParams) (int64, error)
}
