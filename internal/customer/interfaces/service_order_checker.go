package interfaces

import "context"

// ServiceOrderChecker tells whether a customer has service orders. Orders live in the
// service-order service, so the answer comes from its internal REST.
type ServiceOrderChecker interface {
	HasCustomerOrders(ctx context.Context, customerID string) (bool, error)
}
