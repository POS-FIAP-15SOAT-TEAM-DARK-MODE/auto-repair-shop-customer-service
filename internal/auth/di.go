package auth

import (
	"database/sql"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
)

func NewHTTPController(db *sql.DB) interfaces.AuthHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(env.GetTimeDuration("JWT_EXPIRES_IN", 24*time.Hour))
	svc := NewService(unitOfWork, repo, id.NewIDGenerator())
	return NewController(svc)
}
