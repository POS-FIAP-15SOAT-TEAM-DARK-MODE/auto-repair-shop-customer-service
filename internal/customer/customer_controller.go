package customer

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"go.uber.org/zap"
)

type customerController struct {
	svc interfaces.CustomerService
}

func NewController(svc interfaces.CustomerService) *customerController {
	return &customerController{svc}
}

func (ctrl customerController) Create(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error) {
	dto := adapters.CreateCustomerRequest{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("customer.create_payload_invalid",
			zap.String("operation", "create_customer"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	resp, err := ctrl.svc.Create(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("customer.create_failed",
			zap.String("operation", "create_customer"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	return resp, nil
}

func (ctrl customerController) GetByDocument(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error) {
	document := req.URL.Query().Get("document")
	resp, err := ctrl.svc.GetByDocument(ctx, document)
	if err != nil {
		logger.Of(ctx).Debug("customer.get_failed",
			zap.String("operation", "get_by_document"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	return resp, nil
}

func (ctrl customerController) GetByID(ctx context.Context, _ *http.Request) (adapters.CustomerResponse, error) {
	id, _ := ctx.Value("id").(string)
	resp, err := ctrl.svc.GetByID(ctx, id)
	if err != nil {
		logger.Of(ctx).Debug("customer.get_failed",
			zap.String("operation", "get_by_id"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	return resp, nil
}

func (ctrl customerController) Update(ctx context.Context, req *http.Request) (adapters.CustomerResponse, error) {
	dto := adapters.UpdateCustomerRequest{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("customer.update_payload_invalid",
			zap.String("operation", "update_customer"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	id, _ := ctx.Value("id").(string)
	resp, err := ctrl.svc.Update(ctx, id, dto)
	if err != nil {
		logger.Of(ctx).Debug("customer.update_failed",
			zap.String("operation", "update_customer"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return adapters.CustomerResponse{}, err
	}

	return resp, nil
}

func (ctrl customerController) Delete(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value("id").(string)
	if err := ctrl.svc.Delete(ctx, id); err != nil {
		logger.Of(ctx).Debug("customer.delete_failed",
			zap.String("operation", "delete_customer"),
			zap.Error(err),
			zap.String("entity", "customer"),
		)
		return err
	}
	return nil
}
