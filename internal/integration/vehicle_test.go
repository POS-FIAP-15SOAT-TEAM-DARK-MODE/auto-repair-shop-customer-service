//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVehicle_OldFormat(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "TST-1234",
		"brand":        "Fiat",
		"model":        "Uno",
		"year":         2015,
		"customerId":   seedCustomerID,
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	v := decodeJSON[vehicleResponse](t, resp)
	assert.NotEmpty(t, v.ID)
	// plate is stored/returned normalized (no hyphen)
	assert.Equal(t, "TST1234", v.LicensePlate)
}

func TestCreateVehicle_MercosulFormat(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "XYZ2B34",
		"brand":        "Honda",
		"model":        "Civic",
		"year":         2022,
		"customerId":   seedCustomerID,
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	v := decodeJSON[vehicleResponse](t, resp)
	assert.Equal(t, "XYZ2B34", v.LicensePlate)
}

func TestCreateVehicle_InvalidPlate(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "INVALID",
		"brand":        "Ford",
		"model":        "Ka",
		"year":         2019,
		"customerId":   seedCustomerID,
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateVehicle_DuplicatePlate(t *testing.T) {
	plate := "DUP-9999"

	first := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": plate, "brand": "VW", "model": "Golf", "year": 2020, "customerId": seedCustomerID,
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, first.StatusCode)
	first.Body.Close()

	second := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": plate, "brand": "VW", "model": "Polo", "year": 2021, "customerId": seedCustomerID,
	}, attendantToken(t))
	assert.Equal(t, http.StatusConflict, second.StatusCode)
	second.Body.Close()
}

func TestCreateVehicle_MissingBrand(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "MBR-0001", "model": "Model", "year": 2020, "customerId": seedCustomerID,
	}, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateVehicle_MissingCustomer(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "MCU-0001", "brand": "Ford", "model": "Ranger", "year": 2021,
	}, attendantToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestFindVehicleByLicensePlate(t *testing.T) {
	// Seed creates a vehicle with plate "ABC-1234" (stored as "ABC1234").
	resp := doRequest(t, http.MethodGet, "/v1/vehicles?plate=ABC-1234", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	v := decodeJSON[paginatedResponse[vehicleResponse]](t, resp)
	assert.Equal(t, "ABC1234", v.Items[0].LicensePlate)
}

func TestFindVehicleByLicensePlate_NotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/vehicles?plate=ZZZ-9999", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	v := decodeJSON[paginatedResponse[vehicleResponse]](t, resp)
	assert.Len(t, v.Items, 0)
}

func TestUpdateVehicle(t *testing.T) {
	created := createVehicle(t, "UPD-1111", "Renault", "Sandero", 2018, seedCustomerID)

	resp := doRequest(t, http.MethodPut, "/v1/vehicles/"+created.ID, map[string]any{
		"licensePlate": "UPD-1111",
		"brand":        "Renault",
		"model":        "Sandero RS",
		"year":         2019,
		"customerId":   seedCustomerID,
	}, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	updated := decodeJSON[vehicleResponse](t, resp)
	assert.Contains(t, updated.BrandModel, "Sandero RS")
	assert.Equal(t, 2019, updated.Year)
}

func TestDeleteVehicle(t *testing.T) {
	created := createVehicle(t, "DEL-2222", "GM", "Onix", 2021, seedCustomerID)

	resp := doRequest(t, http.MethodDelete, "/v1/vehicles/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestFindVehiclesByCustomer(t *testing.T) {
	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/v1/vehicles?customerId=%s", seedCustomerID), nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	page := decodeJSON[paginatedResponse[vehicleResponse]](t, resp)
	assert.NotEmpty(t, page.Items, "seed customer should have at least one vehicle")
}

func TestCreateVehicle_Unauthorized_Mechanic(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", map[string]any{
		"licensePlate": "MCH-3333", "brand": "T", "model": "T", "year": 2020, "customerId": seedCustomerID,
	}, mechanicToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}
