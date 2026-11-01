package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type repo struct {
	sessionExpirationDuration time.Duration
}

func NewPostgres(sessionExpirationDuration time.Duration) interfaces.AuthRepository {
	return &repo{sessionExpirationDuration: sessionExpirationDuration}
}

func (u *repo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	db, err := uow.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	err = db.QueryRowContext(ctx, getUserByEmail, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrInvalidUserCredentials
		}

		return domain.User{}, pgPkg.Error(ctx, err)
	}

	return user, nil
}

func (u *repo) GetUserSession(ctx context.Context, id string) (adapters.LoginResponse, error) {
	roles, err := u.getRolesByUserId(ctx, id)
	if err != nil {
		return adapters.LoginResponse{}, err
	}
	expiresAt := time.Now().Add(u.sessionExpirationDuration)
	token, err := auth.GenerateToken(id, roles, expiresAt)
	if err != nil {
		return adapters.LoginResponse{}, err
	}

	return adapters.LoginResponse{
		Token:     token,
		ExpiresIn: int64(expiresAt.Second()),
	}, err
}

func (u *repo) getRolesByUserId(ctx context.Context, id string) ([]domain.Role, error) {
	db, err := uow.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	var roles []domain.Role
	rows, err := db.QueryContext(ctx, getRolesById, id)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var role domain.Role
		if err = rows.Scan(&role); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (u *repo) Save(ctx context.Context, c *domain.User) error {
	tx, err := uow.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query",
		zap.String("query", createUserQuery),
		zap.String("user_id", c.ID),
		zap.String("email", c.Email),
	)
	if _, err = tx.ExecContext(ctx, createUserQuery, c.ID, c.Name, c.Email, c.Password); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}

func (u *repo) SaveUserRole(ctx context.Context, userID string, role domain.Role) error {
	tx, err := uow.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, deleteRolesQuery, userID); err != nil {
		return pgPkg.Error(ctx, err)
	}

	id := uuid.New().String()
	if _, err = tx.ExecContext(ctx, updateRoleQuery, id, userID, string(role)); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (u *repo) Update(ctx context.Context, userID, name, email string) error {
	tx, err := uow.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, updateUserQuery, name, email, userID); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (u *repo) Delete(ctx context.Context, userID string) error {
	tx, err := uow.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, deleteRolesQuery, userID); err != nil {
		return pgPkg.Error(ctx, err)
	}

	if _, err = tx.ExecContext(ctx, deleteUserQuery, userID); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}
