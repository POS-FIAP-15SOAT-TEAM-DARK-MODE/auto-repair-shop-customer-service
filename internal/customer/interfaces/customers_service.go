package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerService --with-expecter
type CustomerService interface {
	Create(ctx context.Context, customer adapters.CreateCustomerRequest) (adapters.CustomerResponse, error)
	GetByDocument(ctx context.Context, document string) (adapters.CustomerResponse, error)
	GetByID(ctx context.Context, document string) (adapters.CustomerResponse, error)
	Update(ctx context.Context, id string, customer adapters.UpdateCustomerRequest) (adapters.CustomerResponse, error)
	Delete(ctx context.Context, id string) error
}
