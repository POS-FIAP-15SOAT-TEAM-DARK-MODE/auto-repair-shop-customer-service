package customer

import (
	"context"
	"net/http"
	"strings"
	"unicode"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
)

// GetInternalByID serves GET /internal/customers/{id}. "name" is always the user name;
// "company_name" is set only for COMPANY customers (null otherwise); the document is unformatted.
func (ctrl customerController) GetInternalByID(ctx context.Context, _ *http.Request) (adapters.InternalCustomerResponse, error) {
	id, _ := ctx.Value("id").(string)
	c, err := ctrl.svc.GetByID(ctx, id)
	if err != nil {
		return adapters.InternalCustomerResponse{}, err
	}

	var companyName *string
	if c.Type == domain.CompanyCustomerType.String() {
		name := c.CompanyName
		companyName = &name
	}

	return adapters.InternalCustomerResponse{
		ID:          c.ID,
		UserID:      c.UserResponse.ID,
		Type:        c.Type,
		Name:        c.UserResponse.Name,
		Email:       c.UserResponse.Email,
		Document:    unformatDocument(c.Document),
		CompanyName: companyName,
		Phone:       c.Phone,
	}, nil
}

// unformatDocument keeps only letters and digits (CPF digits, or CNPJ alphanumerics).
func unformatDocument(doc string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, doc)
}
