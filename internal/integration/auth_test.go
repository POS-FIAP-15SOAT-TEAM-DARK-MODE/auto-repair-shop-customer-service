//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_ValidAttendantCredentials(t *testing.T) {
	body := map[string]string{
		"email":    "attendant@autorepairshop.com",
		"password": "attendant123",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeJSON[loginResponse](t, resp)
	assert.NotEmpty(t, result.Token)
	assert.Greater(t, result.ExpiresIn, 0)
}

func TestLogin_ValidMechanicCredentials(t *testing.T) {
	body := map[string]string{
		"email":    "mechanic@autorepairshop.com",
		"password": "mechanic123",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeJSON[loginResponse](t, resp)
	assert.NotEmpty(t, result.Token)
}

func TestLogin_ValidCustomerCredentials(t *testing.T) {
	body := map[string]string{
		"email":    "joao.silva@email.com",
		"password": "529.982.247-25",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeJSON[loginResponse](t, resp)
	assert.NotEmpty(t, result.Token)
}

func TestLogin_WrongPassword(t *testing.T) {
	body := map[string]string{
		"email":    "attendant@autorepairshop.com",
		"password": "wrongpassword",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_UnknownEmail(t *testing.T) {
	body := map[string]string{
		"email":    "nobody@example.com",
		"password": "Test@12345",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_EmptyCredentials(t *testing.T) {
	body := map[string]string{"email": "", "password": ""}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestProtectedEndpoint_NoToken(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/customers", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/customers", nil, "not-a-valid-jwt")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRegisterUser_AsAdmin(t *testing.T) {
	body := map[string]string{
		"name":            "Test Technician",
		"email":           "tech.register@test.com",
		"password":        "Tech@12345",
		"confirmPassword": "Tech@12345",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/register", body, adminToken(t))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	result := decodeJSON[userResponse](t, resp)
	assert.Equal(t, "Test Technician", result.Name)
	assert.Equal(t, "tech.register@test.com", result.Email)
	assert.NotEmpty(t, result.ID)
}

func TestRegisterUser_Unauthorized_NonAdmin(t *testing.T) {
	body := map[string]string{
		"name":            "Sneaky User",
		"email":           "sneaky@test.com",
		"password":        "Sneaky@12345",
		"confirmPassword": "Sneaky@12345",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/register", body, attendantToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRegisterUser_PasswordMismatch(t *testing.T) {
	body := map[string]string{
		"name":            "Test User",
		"email":           "mismatch@test.com",
		"password":        "Test@12345",
		"confirmPassword": "Different@12345",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/register", body, adminToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestUpdateUserRole_AsAdmin(t *testing.T) {
	createBody := map[string]string{
		"name": "Role Update User", "email": "roleupdate@test.com",
		"password": "Test@12345", "confirmPassword": "Test@12345",
	}
	createResp := doRequest(t, http.MethodPost, "/v1/auth/register", createBody, adminToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	user := decodeJSON[userResponse](t, createResp)

	updateBody := map[string]string{"role": "MECHANIC"}
	resp := doRequest(t, http.MethodPatch, "/v1/users/"+user.ID+"/role", updateBody, adminToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestUpdateUserRole_Unauthorized_NonAdmin(t *testing.T) {
	body := map[string]string{"role": "MECHANIC"}
	resp := doRequest(t, http.MethodPatch, "/v1/users/some-id/role", body, attendantToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestLoginResponse_ContainsValidJWT(t *testing.T) {
	body := map[string]string{
		"email":    "attendant@autorepairshop.com",
		"password": "attendant123",
	}
	resp := doRequest(t, http.MethodPost, "/v1/auth/login", body, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	resp.Body.Close()

	protResp := doRequest(t, http.MethodGet, "/v1/works", nil, result.Token)
	assert.Equal(t, http.StatusOK, protResp.StatusCode)
	protResp.Body.Close()
}
