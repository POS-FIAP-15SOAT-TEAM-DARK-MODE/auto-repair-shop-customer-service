package web_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	jsonv2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/web"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	"github.com/stretchr/testify/assert"
)

func TestError_StatusMapping(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"unauthorized", authDomain.ErrInvalidUserCredentials, http.StatusUnauthorized},
		{"forbidden", customerDomain.ErrInvalidCustomerProperty, http.StatusForbidden},
		{"bad request auth", authDomain.ErrEmptyUserName, http.StatusBadRequest},
		{"bad request customer (joined)", errors.Join(customerDomain.ErrInvalidCPF, customerDomain.ErrPhoneRequired), http.StatusBadRequest},
		{"bad request vehicle", vehicleDomain.ErrVehicleInvalidPlate, http.StatusBadRequest},
		{"bad request json", jsonv2.ErrJSONSyntax, http.StatusBadRequest},
		{"not found customer", customerDomain.ErrCustomerNotFound, http.StatusNotFound},
		{"not found vehicle", vehicleDomain.ErrVehicleNotFound, http.StatusNotFound},
		{"conflict", app.ErrDataConflict, http.StatusConflict},
		{"conflict infra", app.ErrInfraConflict, http.StatusConflict},
		{"conflict customer", customerDomain.ErrCustomerHasServiceOrders, http.StatusConflict},
		{"unprocessable", app.ErrDataViolation, http.StatusUnprocessableEntity},
		{"internal json", jsonv2.ErrJSONUnsupportedValue, http.StatusInternalServerError},
		{"unknown", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := web.Error(tc.err)
			assert.Equal(t, tc.status, status)
			assert.Equal(t, tc.status, body.StatusCode)
		})
	}
}

func TestError_BodyShape(t *testing.T) {
	_, body := web.Error(customerDomain.ErrCustomerNotFound)
	assert.Nil(t, body.Errors)
	assert.Equal(t, "customer not found", body.Error)

	_, body = web.Error(fmt.Errorf("wrap: %w", app.ErrDataConflict))
	assert.Equal(t, []string(nil), body.Errors)
	assert.Equal(t, "conflict", body.Error)

	joined := errors.Join(authDomain.ErrEmptyUserName, authDomain.ErrInvalidUserEmail)
	status, body := web.Error(joined)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Len(t, body.Errors, 2)
	assert.Empty(t, body.Error)

	_, body = web.Error(errors.New("boom"))
	assert.Equal(t, "internal server error", body.Error)
}

// TestError_NeverLeaksInfrastructureCause proves that the text of an infrastructure cause (driver,
// constraint, document value) never reaches the body, for every status that web.Error can produce.
func TestError_NeverLeaksInfrastructureCause(t *testing.T) {
	const sensitive = `pq: duplicate key value violates unique constraint "customers_document_key" (52998224725)`
	cause := errors.New(sensitive)

	cases := []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{"409 infra", fmt.Errorf("insert %s: %w", sensitive, app.ErrDataConflict), http.StatusConflict, "conflict"},
		{"409 infra retry", fmt.Errorf("tx: %w: %s", app.ErrInfraConflict, sensitive), http.StatusConflict, "conflict"},
		{"422 infra", fmt.Errorf("insert %s: %w", sensitive, app.ErrDataViolation), http.StatusUnprocessableEntity, "invalid data"},
		{"500 driver", fmt.Errorf("query: %w", cause), http.StatusInternalServerError, "internal server error"},
		{"500 json", fmt.Errorf("encode: %w", jsonv2.ErrJSONUnsupportedValue), http.StatusInternalServerError, "internal server error"},
		{"500 unknown", errors.New(sensitive), http.StatusInternalServerError, "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := web.Error(tc.err)
			assert.Equal(t, tc.status, status)
			assert.Equal(t, tc.status, body.StatusCode)
			assert.Equal(t, tc.want, body.Error)
			assert.Nil(t, body.Errors)
			assert.NotContains(t, body.Error, "pq")
			assert.NotContains(t, body.Error, "52998224725")
		})
	}
}

// TestError_DomainMessageIsKept proves that known domain errors keep the message of their own sentinel,
// without the text of the wrapper that carries them.
func TestError_DomainMessageIsKept(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{"404 customer", fmt.Errorf("find 52998224725: %w", customerDomain.ErrCustomerNotFound), http.StatusNotFound, "customer not found"},
		{"404 vehicle", fmt.Errorf("find ABC1D23: %w", vehicleDomain.ErrVehicleNotFound), http.StatusNotFound, "vehicle not found"},
		{"409 service orders", fmt.Errorf("delete 7c9e: %w", customerDomain.ErrCustomerHasServiceOrders), http.StatusConflict, "customer has associated service orders"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := web.Error(tc.err)
			assert.Equal(t, tc.status, status)
			assert.Equal(t, tc.want, body.Error)
			assert.Nil(t, body.Errors)
		})
	}
}

func TestIsInfrastructure(t *testing.T) {
	assert.True(t, web.IsInfrastructure(fmt.Errorf("x: %w", app.ErrDataConflict)))
	assert.True(t, web.IsInfrastructure(app.ErrInfraConflict))
	assert.True(t, web.IsInfrastructure(app.ErrDataViolation))
	assert.False(t, web.IsInfrastructure(customerDomain.ErrCustomerHasServiceOrders))
	assert.False(t, web.IsInfrastructure(errors.New("boom")))
}

func TestIsJoined(t *testing.T) {
	assert.True(t, web.IsJoined(errors.Join(errors.New("a"), errors.New("b"))))
	assert.False(t, web.IsJoined(errors.New("a")))
}

func TestError_InvalidDocumentIsBadRequest(t *testing.T) {
	for _, err := range []error{
		customerDomain.ErrInvalidCPF,
		customerDomain.ErrCPFLength,
		customerDomain.ErrInvalidCNPJ,
		customerDomain.ErrCNPJLength,
	} {
		status, _ := web.Error(err)
		assert.Equal(t, http.StatusBadRequest, status, err.Error())
	}
}
