package adapters

// InternalCustomerResponse is the payload of GET /internal/customers/{id}, consumed by service-order.
// Contract: docs/fase4/contracts/internal-rest.md.
type InternalCustomerResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Document    string  `json:"document"`
	CompanyName *string `json:"company_name"`
	Phone       string  `json:"phone"`
}
