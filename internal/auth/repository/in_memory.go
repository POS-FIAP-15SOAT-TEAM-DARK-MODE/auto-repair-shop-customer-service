package repository

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
)

type memory_repo struct {
	data  map[string]domain.User
	roles map[string][]domain.Role
}

func NewInMemory() interfaces.AuthRepository {
	return &memory_repo{
		data:  make(map[string]domain.User),
		roles: make(map[string][]domain.Role),
	}
}

func (r *memory_repo) Save(_ context.Context, u *domain.User) error {
	for _, existing := range r.data {
		if existing.Email == u.Email {
			return app.ErrDataConflict
		}
	}
	r.data[u.ID] = *u
	return nil
}

func (r *memory_repo) SaveUserRole(_ context.Context, userID string, role domain.Role) error {
	r.roles[userID] = []domain.Role{role}
	return nil
}

func (r *memory_repo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	for _, u := range r.data {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, nil
}

func (r *memory_repo) GetUserSession(_ context.Context, userID string) (adapters.LoginResponse, error) {
	roles := r.roles[userID]
	expiresAt := time.Now().Add(time.Hour)
	token, err := auth.GenerateToken(userID, roles, expiresAt)
	if err != nil {
		return adapters.LoginResponse{}, err
	}

	return adapters.LoginResponse{
		Token:     token,
		ExpiresIn: int64(expiresAt.Second()),
	}, err
}

func (r *memory_repo) Update(_ context.Context, userID, name, email string) error {
	u, ok := r.data[userID]
	if !ok {
		return domain.ErrInvalidUserCredentials
	}
	u.Name = name
	u.Email = email
	r.data[userID] = u
	return nil
}

func (r *memory_repo) Delete(_ context.Context, userID string) error {
	if _, ok := r.data[userID]; !ok {
		return domain.ErrInvalidUserCredentials
	}
	delete(r.data, userID)
	delete(r.roles, userID)
	return nil
}
