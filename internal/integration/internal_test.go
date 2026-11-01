//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// internalGet calls an /internal route. token == "" sends no X-Internal-Token header.
func internalGet(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, testServer.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("X-Internal-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestInternalCustomer_OK(t *testing.T) {
	resp := internalGet(t, "/internal/customers/"+seedCustomerID, internalToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, seedCustomerID, body["id"])
	assert.Equal(t, seedUserID, body["user_id"])
	assert.Equal(t, "INDIVIDUAL", body["type"])
	assert.Equal(t, "João Silva", body["name"])
	assert.Equal(t, "52998224725", body["document"])
}

func TestInternalCustomer_NotFound(t *testing.T) {
	resp := internalGet(t, "/internal/customers/"+uuid.NewString(), internalToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInternalCustomer_Unauthorized(t *testing.T) {
	assert.Equal(t, http.StatusUnauthorized, internalGet(t, "/internal/customers/"+seedCustomerID, "").StatusCode)
	assert.Equal(t, http.StatusUnauthorized, internalGet(t, "/internal/customers/"+seedCustomerID, "wrong").StatusCode)
}

func TestInternalVehicle_OK(t *testing.T) {
	resp := internalGet(t, "/internal/vehicles/"+seedVehicleID, internalToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, seedVehicleID, body["id"])
	assert.Equal(t, seedCustomerID, body["customer_id"])
	assert.Equal(t, "ABC1234", body["license_plate"])
	assert.Equal(t, "Toyota", body["brand"])
	assert.Equal(t, "Corolla", body["model"])
	assert.Equal(t, float64(2020), body["year"])
}

func TestInternalVehicle_NotFound(t *testing.T) {
	resp := internalGet(t, "/internal/vehicles/"+uuid.NewString(), internalToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestInternalVehicle_Unauthorized(t *testing.T) {
	assert.Equal(t, http.StatusUnauthorized, internalGet(t, "/internal/vehicles/"+seedVehicleID, "").StatusCode)
}
