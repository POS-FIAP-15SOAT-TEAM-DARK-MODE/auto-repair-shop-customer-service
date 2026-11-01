package adapters

type CreateCustomerRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Document    string `json:"document"`
	CompanyName string `json:"companyName,omitempty"`
	Phone       string `json:"phone"`
}

type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
