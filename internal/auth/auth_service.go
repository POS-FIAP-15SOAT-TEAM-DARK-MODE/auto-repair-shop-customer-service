package auth

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/uow"
	"go.uber.org/zap"
)

type service struct {
	uow         uow.Executor
	repo        interfaces.AuthRepository
	idGenerator id.Generator
}

func NewService(uow uow.Executor, repo interfaces.AuthRepository, idGenerator id.Generator) interfaces.AuthService {
	return &service{
		uow:         uow,
		repo:        repo,
		idGenerator: idGenerator,
	}
}

func (s service) Login(ctx context.Context, data adapters.LoginRequest) (adapters.LoginResponse, error) {
	if err := data.Domain().CheckRequiredFields(); err != nil {
		return adapters.LoginResponse{}, err
	}

	user, err := s.repo.GetByEmail(ctx, data.Email)
	if err != nil {
		return adapters.LoginResponse{}, err
	}

	if err = user.ComparePassword(data.Password); err != nil {
		return adapters.LoginResponse{}, err
	}

	return s.repo.GetUserSession(ctx, user.ID)
}

func (s service) Register(ctx context.Context, data adapters.CreateUserRequest) (adapters.UserResponse, error) {
	if data.Password != data.ConfirmPassword {
		return adapters.UserResponse{}, domain.ErrUserPasswordDontMatch
	}

	user := domain.NewUser(
		s.idGenerator.NewUUID(),
		data.Name,
		data.Email,
		data.Password,
	)

	if err := user.Validate(); err != nil {
		return adapters.UserResponse{}, err
	}

	if err := user.HashPassword(ctx); err != nil {
		return adapters.UserResponse{}, err
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.repo.Save(ctx, user); err != nil {
			return err
		}

		return s.repo.SaveUserRole(ctx, user.ID, domain.CUSTOMER)
	})
	if err != nil {
		return adapters.UserResponse{}, err
	}

	return adapters.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s service) ChangeRole(ctx context.Context, userId string, data adapters.ChangeRoleRequest) error {
	if err := data.IsValid(); err != nil {
		return err
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.repo.SaveUserRole(ctx, userId, data.Role)
	})
	if err != nil {
		return err
	}

	logger.Of(ctx).Info("user.role_changed", zap.String("user_id", userId), zap.String("role", string(data.Role)))
	return nil
}
