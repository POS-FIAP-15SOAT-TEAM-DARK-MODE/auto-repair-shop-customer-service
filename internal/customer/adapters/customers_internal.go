package adapters

// InternalCustomerResponse is the payload of GET /internal/customers/{id}, consumed by service-order.
type InternalCustomerResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Document string `json:"document"`
}
