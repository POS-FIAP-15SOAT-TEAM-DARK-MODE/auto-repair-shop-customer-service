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

// TestError_ValidationSentinelsAnswer400 proves that a validation sentinel anywhere in the error tree
// answers 400, and that the errors list carries only the matched sentinels' messages.
func TestError_ValidationSentinelsAnswer400(t *testing.T) {
	driverErr := errors.New(`pq: duplicate key value violates unique constraint "customers_document_key"`)
	decimalErr := errors.New("decimal: cannot parse 'abc' as a number")

	t.Run("join of validation sentinels lists each message", func(t *testing.T) {
		status, body := web.Error(errors.Join(customerDomain.ErrInvalidCPF, customerDomain.ErrPhoneRequired))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF", "phone is required"}, body.Errors)
		assert.Empty(t, body.Error)
	})

	t.Run("wrapped join of validation sentinels has no wrapper text", func(t *testing.T) {
		status, body := web.Error(fmt.Errorf("create customer 52998224725: %w",
			errors.Join(customerDomain.ErrInvalidCPF, customerDomain.ErrCPFLength)))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF", "CPF must have 11 digits"}, body.Errors)
	})

	t.Run("validation with a driver cause is 400 with only the sentinel text", func(t *testing.T) {
		status, body := web.Error(errors.Join(customerDomain.ErrInvalidCPF, driverErr))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF"}, body.Errors)
		assert.NotContains(t, body.Errors, driverErr.Error())
	})

	t.Run("validation with a domain cause that is not validation is 400", func(t *testing.T) {
		status, body := web.Error(errors.Join(customerDomain.ErrInvalidCPF, customerDomain.ErrCustomerNotFound))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF"}, body.Errors)
	})

	t.Run("validation with an unknown cause in a %w wrap is 400 with only the sentinel text", func(t *testing.T) {
		status, body := web.Error(fmt.Errorf("%w: %w", customerDomain.ErrInvalidCPF, decimalErr))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF"}, body.Errors)
		assert.NotContains(t, body.Errors, decimalErr.Error())
	})

	t.Run("document error together with a parsing error keeps both messages", func(t *testing.T) {
		status, body := web.Error(errors.Join(customerDomain.ErrInvalidCPF, jsonv2.ErrJSONSyntax))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF", "invalid JSON syntax"}, body.Errors)
	})

	t.Run("single validation sentinel wrapped keeps only the sentinel message", func(t *testing.T) {
		status, body := web.Error(fmt.Errorf("document 123: %w", customerDomain.ErrInvalidCPF))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"invalid CPF"}, body.Errors)
	})

	t.Run("vehicle without customer is 400", func(t *testing.T) {
		status, body := web.Error(vehicleDomain.ErrVehicleNoCustomerAssociated)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, []string{"no customer associated to the vehicle"}, body.Errors)
	})
}

// TestError_InfrastructureOrUnknownOnlyIsGeneric proves that errors without any validation sentinel
// never reach the body: they answer a fixed message for their status.
func TestError_InfrastructureOrUnknownOnlyIsGeneric(t *testing.T) {
	driverErr := errors.New(`pq: duplicate key value violates unique constraint "customers_document_key"`)
	decimalErr := errors.New("decimal: cannot parse 'abc' as a number")

	t.Run("join of driver and unknown errors is 500", func(t *testing.T) {
		status, body := web.Error(errors.Join(driverErr, decimalErr))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Nil(t, body.Errors)
		assert.Equal(t, "internal server error", body.Error)
		assert.NotContains(t, body.Error, "pq")
	})

	t.Run("wrapped join of unknown errors is 500", func(t *testing.T) {
		status, body := web.Error(fmt.Errorf("save: %w", errors.Join(driverErr, decimalErr)))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "internal server error", body.Error)
	})

	t.Run("join of a conflict and a driver error is 409 with the generic message", func(t *testing.T) {
		status, body := web.Error(errors.Join(app.ErrDataConflict, driverErr))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "conflict", body.Error)
		assert.NotContains(t, body.Error, "pq")
	})
}

// TestError_DomainAuthErrorsUseErrorString: domain 401 and 403 answer error (string), never the errors list.
func TestError_DomainAuthErrorsUseErrorString(t *testing.T) {
	status, body := web.Error(authDomain.ErrInvalidUserCredentials)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Nil(t, body.Errors)
	assert.Equal(t, "invalid user credentials", body.Error)

	status, body = web.Error(customerDomain.ErrInvalidCustomerProperty)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Nil(t, body.Errors)
	assert.Equal(t, "customer not allowed to execute this operation", body.Error)
}

func TestIsInfrastructure(t *testing.T) {
	assert.True(t, web.IsInfrastructure(fmt.Errorf("x: %w", app.ErrDataConflict)))
	assert.True(t, web.IsInfrastructure(app.ErrInfraConflict))
	assert.True(t, web.IsInfrastructure(app.ErrDataViolation))
	assert.False(t, web.IsInfrastructure(customerDomain.ErrCustomerHasServiceOrders))
	assert.False(t, web.IsInfrastructure(errors.New("boom")))
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
