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
	assert.Equal(t, "resource already exists", body.Error)

	joined := errors.Join(authDomain.ErrEmptyUserName, authDomain.ErrInvalidUserEmail)
	status, body := web.Error(joined)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Len(t, body.Errors, 2)
	assert.Empty(t, body.Error)

	_, body = web.Error(errors.New("boom"))
	assert.Equal(t, "boom", body.Error)
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
