package auth

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"go.uber.org/zap"
)

type authController struct {
	svc interfaces.AuthService
}

func NewController(svc interfaces.AuthService) *authController {
	return &authController{svc}
}

func (ctrl authController) Login(ctx context.Context, req *http.Request) (adapters.LoginResponse, error) {
	dto := adapters.LoginRequest{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind login payload",
			zap.String("operation", "login"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return adapters.LoginResponse{}, err
	}

	resp, err := ctrl.svc.Login(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("Failed to register user",
			zap.String("operation", "log_in_user"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return adapters.LoginResponse{}, err
	}

	return resp, nil
}

func (ctrl authController) Register(ctx context.Context, req *http.Request) (adapters.UserResponse, error) {
	dto := adapters.CreateUserRequest{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind register payload",
			zap.String("operation", "register"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return adapters.UserResponse{}, err
	}

	resp, err := ctrl.svc.Register(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("Failed to register user",
			zap.String("operation", "register_user"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return adapters.UserResponse{}, err
	}

	return resp, nil
}

func (ctrl authController) ChangeRole(ctx context.Context, req *http.Request) error {
	dto := adapters.ChangeRoleRequest{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind change role payload",
			zap.String("operation", "change_user_role"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return err
	}

	id, _ := ctx.Value("id").(string)
	if err := ctrl.svc.ChangeRole(ctx, id, dto); err != nil {
		logger.Of(ctx).Debug("Failed to change user roles",
			zap.String("operation", "change_user_role"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		return err
	}

	return nil
}
