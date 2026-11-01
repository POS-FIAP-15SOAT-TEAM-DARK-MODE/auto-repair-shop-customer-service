package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
)

type memoryRepository struct {
	data map[string]domain.Customer
}

func NewInMemory() interfaces.CustomerRepository {
	return &memoryRepository{
		data: make(map[string]domain.Customer),
	}
}

func (r *memoryRepository) Save(_ context.Context, c *domain.Customer) error {
	for id, existing := range r.data {
		if id == c.ID {
			continue
		}
		if (c.CPF != "" && existing.CPF == c.CPF) || (c.CNPJ != "" && existing.CNPJ == c.CNPJ) {
			return app.ErrDataConflict
		}
	}
	stored := *c
	if c.User != nil {
		u := *c.User
		stored.User = &u
	}
	r.data[c.ID] = stored
	return nil
}

func (r *memoryRepository) Delete(_ context.Context, id string) error {
	if _, ok := r.data[id]; !ok {
		return domain.ErrCustomerNotFound
	}
	delete(r.data, id)
	return nil
}

func (r *memoryRepository) List(_ context.Context, params adapters.ListCustomerParams) ([]domain.Customer, error) {
	var results []domain.Customer

	for _, c := range r.data {
		if params.ID != "" && c.ID != params.ID {
			continue
		}
		if params.Document != "" && c.CPF != params.Document && c.CNPJ != params.Document {
			continue
		}
		if params.UserID != "" && c.UserID != params.UserID {
			continue
		}

		stored := c
		if c.User != nil {
			u := authDomain.User{
				ID:   c.User.ID,
				Name: c.User.Name,
				CoreUser: authDomain.CoreUser{
					Email: c.User.Email,
				},
			}
			stored.User = &u
		}
		results = append(results, stored)
	}

	return results, nil
}

var _ interfaces.CustomerRepository = (*memoryRepository)(nil)
