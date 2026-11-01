package vehicle

import (
	"context"
	"net/http"
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces"
	"go.uber.org/zap"
)

const (
	plateParam      = "plate"
	idParam         = "id"
	customerIdParam = "customerId"
	pageParam       = "page"
	pageSizeParam   = "pageSize"
	defaultPageSize = 10
	defaultPage     = 1
)

type vehicleController struct {
	svc interfaces.VehicleService
}

func NewController(svc interfaces.VehicleService) *vehicleController {
	return &vehicleController{
		svc: svc,
	}
}

func (ctrl *vehicleController) Create(ctx context.Context, req *http.Request) (adapters.VehicleResponse, error) {
	dto := adapters.CreateVehicle{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Vehicle creation error",
			zap.String("operation", "create_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return adapters.VehicleResponse{}, err
	}

	resp, err := ctrl.svc.Create(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("Vehicle creation error",
			zap.String("operation", "create_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return adapters.VehicleResponse{}, err
	}

	return resp, nil
}

func (ctrl *vehicleController) List(ctx context.Context, req *http.Request) (adapters.PaginatedVehicleResponse, error) {
	params := req.URL.Query()
	pg := int64(defaultPage)
	pgSize := int64(defaultPageSize)
	if page := params.Get(pageParam); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > defaultPage {
			pg = int64(p)
		}
	}
	if pageSize := params.Get(pageSizeParam); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pgSize = int64(ps)
		}
	}

	result, err := ctrl.svc.List(ctx, adapters.ListVehiclesParams{
		CustomerID: params.Get(customerIdParam),
		Plate:      params.Get(plateParam),
		Page:       pg,
		PageSize:   pgSize,
	})
	if err != nil {
		logger.Of(ctx).Debug("get vehicles error",
			zap.String("operation", "list_vehicles"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return result, err
	}
	return result, nil
}

func (ctrl *vehicleController) Edit(ctx context.Context, req *http.Request) (adapters.VehicleResponse, error) {
	dto := adapters.CreateVehicle{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Vehicle validation error",
			zap.String("operation", "update_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return adapters.VehicleResponse{}, err
	}

	id, _ := ctx.Value(idParam).(string)
	res, err := ctrl.svc.Edit(ctx, id, dto)
	if err != nil {
		logger.Of(ctx).Debug("Vehicle validation error",
			zap.String("operation", "update_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return adapters.VehicleResponse{}, err
	}

	return res, nil
}

func (ctrl *vehicleController) Delete(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(idParam).(string)
	if err := ctrl.svc.Delete(ctx, id); err != nil {
		logger.Of(ctx).Debug("Vehicle delete error",
			zap.String("operation", "delete_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		return err
	}

	return nil
}
