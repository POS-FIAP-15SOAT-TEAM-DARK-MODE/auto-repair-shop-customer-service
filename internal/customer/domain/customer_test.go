package domain_test

import (
	"testing"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validCPF      = "529.982.247-25"
	validCPFRaw   = "52998224725"
	validCNPJ     = "45.997.418/0001-53"
	validCNPJRaw  = "45997418000153"
	invalidCPF    = "529.982.247-26"
	invalidCNPJ   = "11.222.333/0001-82"
	sameDigitsCPF = "111.111.111-11"
)

func TestNewCustomer_Individual(t *testing.T) {
	c, err := domain.NewCustomer("João Silva", "joao@email.com", domain.IndividualCustomerType, validCPF, "", "11999990001")
	require.NoError(t, err)
	assert.Equal(t, domain.IndividualCustomerType, c.Type)
	assert.Equal(t, validCPFRaw, c.CPF)
	assert.Empty(t, c.CNPJ)
	assert.Equal(t, c.ID, c.UserID)
	assert.Equal(t, c.ID, c.User.ID)
	assert.Equal(t, "INDIVIDUAL", c.Type.String())
}

func TestNewCustomer_Company(t *testing.T) {
	c, err := domain.NewCustomer("Auto Transportes", "contato@t.com.br", domain.CompanyCustomerType, validCNPJ, "Transportes Ltda", "11988880002")
	require.NoError(t, err)
	assert.Equal(t, validCNPJRaw, c.CNPJ)
	assert.Empty(t, c.CPF)
	assert.Equal(t, "Transportes Ltda", c.CompanyName)
}

func TestNewCustomer_Validation(t *testing.T) {
	_, err := domain.NewCustomer("João Silva", "joao@email.com", domain.IndividualCustomerType, validCPF, "", "")
	assert.ErrorIs(t, err, domain.ErrPhoneRequired)

	_, err = domain.NewCustomer("João Silva", "joao@email.com", domain.IndividualCustomerType, validCPF, "Empresa", "1199")
	assert.ErrorIs(t, err, domain.ErrCompanyNameNotAllowed)

	_, err = domain.NewCustomer("Auto Transportes", "contato@t.com", domain.CompanyCustomerType, validCNPJ, "", "1199")
	assert.ErrorIs(t, err, domain.ErrCompanyNameRequired)

	_, err = domain.NewCustomer("João Silva", "joao@email.com", domain.CustomerType("OTHER"), validCPF, "", "1199")
	assert.ErrorIs(t, err, domain.ErrInvalidCustomerType)

	_, err = domain.NewCustomer("João Silva", "joao@email.com", domain.IndividualCustomerType, "123", "", "1199")
	assert.ErrorIs(t, err, domain.ErrCPFLength)

	_, err = domain.NewCustomer("Auto Transportes", "contato@t.com", domain.CompanyCustomerType, "123", "Ltda", "1199")
	assert.ErrorIs(t, err, domain.ErrCNPJLength)

	_, err = domain.NewCustomer("Jo", "joao@email.com", domain.IndividualCustomerType, validCPF, "", "1199")
	assert.ErrorIs(t, err, authDomain.ErrInvalidUserName)
}

func TestParseDocument(t *testing.T) {
	typ, doc, err := domain.ParseDocument(validCPF)
	require.NoError(t, err)
	assert.Equal(t, domain.IndividualCustomerType, typ)
	assert.Equal(t, validCPFRaw, doc)

	typ, doc, err = domain.ParseDocument(validCNPJ)
	require.NoError(t, err)
	assert.Equal(t, domain.CompanyCustomerType, typ)
	assert.Equal(t, validCNPJRaw, doc)

	_, _, err = domain.ParseDocument("1234")
	assert.ErrorIs(t, err, domain.ErrInvalidDocumentFormat)

	_, _, err = domain.ParseDocument(invalidCPF)
	assert.ErrorIs(t, err, domain.ErrInvalidCPF)

	_, _, err = domain.ParseDocument(sameDigitsCPF)
	assert.ErrorIs(t, err, domain.ErrInvalidCPF)

	_, _, err = domain.ParseDocument(invalidCNPJ)
	assert.ErrorIs(t, err, domain.ErrInvalidCNPJ)

	_, _, err = domain.ParseDocument("11111111111111")
	assert.ErrorIs(t, err, domain.ErrInvalidCNPJ)
}

func TestCustomerValidate_CNPJCheckDigitsAreNumeric(t *testing.T) {
	c := &domain.Customer{Type: domain.CompanyCustomerType, CNPJ: "1122233300018X", CompanyName: "Ltda", Phone: "1"}
	assert.ErrorIs(t, c.Validate(), domain.ErrInvalidCNPJ)
}
