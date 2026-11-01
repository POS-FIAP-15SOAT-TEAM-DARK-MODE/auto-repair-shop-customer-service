package vehicle

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
)

// GetInternalByID serves GET /internal/vehicles/{id}.
func (ctrl *vehicleController) GetInternalByID(ctx context.Context, _ *http.Request) (adapters.InternalVehicleResponse, error) {
	id, _ := ctx.Value(idParam).(string)
	v, err := ctrl.svc.FindById(ctx, id)
	if err != nil {
		return adapters.InternalVehicleResponse{}, err
	}
	return adapters.VehicleDomainToInternalResponse(v), nil
}
