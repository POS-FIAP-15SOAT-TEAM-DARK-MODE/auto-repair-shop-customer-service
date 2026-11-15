package interfaces

import "context"

// ServiceOrderChecker tells whether a vehicle has service orders. Orders live in the
// service-order service, so the answer comes from its internal REST.
type ServiceOrderChecker interface {
	HasVehicleOrders(ctx context.Context, vehicleID string) (bool, error)
}
