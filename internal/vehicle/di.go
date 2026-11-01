package vehicle

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/repository"
)

func NewHTTPController(db *sql.DB) interfaces.VehicleHTTPController {
	return newVehicleController(db)
}

// NewInternalHTTPController builds the controller of the /internal/vehicles routes.
func NewInternalHTTPController(db *sql.DB) interfaces.VehicleInternalHTTPController {
	return newVehicleController(db)
}

func newVehicleController(db *sql.DB) *vehicleController {
	unitOfWork := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres()
	svc := NewService(unitOfWork, repo, id.NewIDGenerator())
	return NewController(svc)
}
