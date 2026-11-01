package customer

import (
	"database/sql"
	"time"

	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
)

func NewHTTPController(db *sql.DB) interfaces.CustomerHTTPController {
	return newCustomerController(db)
}

// NewInternalHTTPController builds the controller of the /internal/customers routes.
func NewInternalHTTPController(db *sql.DB) interfaces.CustomerInternalHTTPController {
	return newCustomerController(db)
}

func newCustomerController(db *sql.DB) *customerController {
	unitOfWork := uow.NewTransactionalUoW(db)
	custRepo := customerRepo.NewPostgres()
	userRepo := authRepo.NewPostgres(time.Duration(0)) // session duration unused by customer service
	svc := NewService(unitOfWork, custRepo, userRepo)
	return NewController(svc)
}
