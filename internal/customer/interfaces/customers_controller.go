package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerHTTPController --with-expecter
type CustomerHTTPController interface {
	Create(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error)
	GetByDocument(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error)
	GetByID(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error)
	Update(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error)
	Delete(ctx context.Context, req *http.Request) error
}

// CustomerInternalHTTPController serves the service-to-service routes under /internal.
type CustomerInternalHTTPController interface {
	GetInternalByID(ctx context.Context, req *http.Request) (adapters.InternalCustomerResponse, error)
}
