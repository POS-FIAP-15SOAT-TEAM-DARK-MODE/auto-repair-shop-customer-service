package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerRepository --with-expecter
type CustomerRepository interface {
	Save(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params adapters.ListCustomerParams) ([]domain.Customer, error)
}
