//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCustomer_Individual(t *testing.T) {
	cpf := nextCPF()
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name":     "Ana Pereira",
		"email":    nextEmail(),
		"password": "Test@12345",
		"type":     "INDIVIDUAL",
		"document": cpf,
		"phone":    nextPhone(),
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	c := decodeJSON[customerResponse](t, resp)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Ana Pereira", c.Name)
	assert.Equal(t, "INDIVIDUAL", c.Type)
	assert.Equal(t, cpf, c.Document)
}

func TestCreateCustomer_Company(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name":        "Empresa Teste",
		"email":       nextEmail(),
		"password":    "Test@12345",
		"type":        "COMPANY",
		"document":    cnpj1,
		"companyName": "Empresa Teste LTDA",
		"phone":       nextPhone(),
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	c := decodeJSON[customerResponse](t, resp)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "COMPANY", c.Type)
	assert.Equal(t, cnpj1Sanitized, c.Document)
	assert.Equal(t, "Empresa Teste LTDA", c.CompanyName)
}

func TestCreateCustomer_DuplicateCPF(t *testing.T) {
	sharedCPF := nextCPF()

	first := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Primeiro Cliente", "email": nextEmail(),
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": sharedCPF, "phone": nextPhone(),
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, first.StatusCode)
	first.Body.Close()

	second := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Segundo Cliente", "email": nextEmail(),
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": sharedCPF, "phone": nextPhone(),
	}, attendantToken(t))
	assert.Equal(t, http.StatusConflict, second.StatusCode)
	second.Body.Close()
}

func TestCreateCustomer_DuplicateEmail(t *testing.T) {
	sharedEmail := nextEmail()

	first := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Email Owner", "email": sharedEmail,
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": nextCPF(), "phone": nextPhone(),
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, first.StatusCode)
	first.Body.Close()

	second := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Dup Email", "email": sharedEmail,
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": nextCPF(), "phone": nextPhone(),
	}, attendantToken(t))
	assert.Equal(t, http.StatusConflict, second.StatusCode)
	second.Body.Close()
}

func TestCreateCustomer_InvalidCPF(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Bad CPF", "email": nextEmail(),
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": "000.000.000-00", "phone": nextPhone(),
	}, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateCustomer_InvalidCNPJ(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Bad CNPJ", "email": nextEmail(),
		"password": "Test@12345", "type": "COMPANY",
		"document": "00.000.000/0000-00", "companyName": "Bad Corp",
		"phone": nextPhone(),
	}, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateCustomer_MissingDocument(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "No Doc", "email": nextEmail(),
		"password": "Test@12345", "type": "INDIVIDUAL", "phone": nextPhone(),
	}, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateCustomer_Unauthorized_Mechanic(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/customers", map[string]any{
		"name": "Unauthorized", "email": nextEmail(),
		"password": "Test@12345", "type": "INDIVIDUAL",
		"document": nextCPF(), "phone": nextPhone(),
	}, mechanicToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestGetCustomerByID(t *testing.T) {
	created := createCustomer(t, "Get By ID", nextEmail(), nextCPF(), nextPhone())

	resp := doRequest(t, http.MethodGet, "/v1/customers/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fetched := decodeJSON[customerResponse](t, resp)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Get By ID", fetched.Name)
}

func TestGetCustomerByID_NotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/customers/00000000-0000-0000-0000-000000000000", nil, attendantToken(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestGetCustomerByDocument(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/customers?document=529.982.247-25", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	c := decodeJSON[customerResponse](t, resp)
	assert.Equal(t, "52998224725", c.Document)
}

func TestUpdateCustomer(t *testing.T) {
	created := createCustomer(t, "Update Me", nextEmail(), nextCPF(), nextPhone())

	newName := "Updated Name"
	newPhone := nextPhone()
	resp := doRequest(t, http.MethodPut, "/v1/customers/"+created.ID, map[string]any{
		"name":  newName,
		"phone": newPhone,
	}, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updated := decodeJSON[customerResponse](t, resp)
	assert.Equal(t, newName, updated.Name)
}

func TestDeleteCustomer(t *testing.T) {
	created := createCustomer(t, "Delete Me", nextEmail(), nextCPF(), nextPhone())

	resp := doRequest(t, http.MethodDelete, "/v1/customers/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	getResp := doRequest(t, http.MethodGet, "/v1/customers/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	getResp.Body.Close()
}

func TestDeleteCustomer_NotFound_ReturnsNoContent(t *testing.T) {
	resp := doRequest(t, http.MethodDelete, "/v1/customers/00000000-0000-0000-0000-000000000001", nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestGetCustomerByDocument_MissingQueryParam(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/customers", nil, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
