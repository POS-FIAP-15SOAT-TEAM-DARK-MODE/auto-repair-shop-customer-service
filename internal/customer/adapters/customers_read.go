package adapters

import (
	userAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
)

type CustomerResponse struct {
	userAdapters.UserResponse
	ID          string `json:"id"`
	Type        string `json:"type"`
	Document    string `json:"document"`
	CompanyName string `json:"companyName"`
	Phone       string `json:"phone"`
}

type ListCustomerParams struct {
	ID       string
	Document string
	UserID   string
}
