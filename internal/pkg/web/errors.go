package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	jsonv2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/json"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
)

type errorResponse struct {
	StatusCode int      `json:"code"`
	Errors     []string `json:"errors,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Error builds an errorResponse for the given error and decides the status code.
func Error(err error) (int, errorResponse) {
	status := getHTTPStatus(err)
	stackedErrors, errorMsg := buildErrorMessages(err, status)

	return status, errorResponse{
		StatusCode: status,
		Errors:     stackedErrors,
		Error:      errorMsg,
	}
}

// getHTTPStatus determines the HTTP status code for a given error.
func getHTTPStatus(err error) int {
	switch {
	case isUnauthorizedError(err):
		return http.StatusUnauthorized
	case isForbidden(err):
		return http.StatusForbidden
	case isBadRequestError(err), IsJoined(err):
		return http.StatusBadRequest
	case isNotFoundError(err):
		return http.StatusNotFound
	case isInternalServerError(err):
		return http.StatusInternalServerError
	case isConflictError(err):
		return http.StatusConflict
	case isUnprocessableEntityError(err):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func IsJoined(err error) bool {
	type unwrapper interface {
		Unwrap() []error
	}

	var uw unwrapper
	return errors.As(err, &uw)
}

// notFoundErrors and conflictDomainErrors are the domain sentinels whose messages may be returned to the client.
var (
	notFoundErrors       = []error{vehicleDomain.ErrVehicleNotFound, customerDomain.ErrCustomerNotFound}
	conflictDomainErrors = []error{customerDomain.ErrCustomerHasServiceOrders}
)

// IsInfrastructure reports whether err comes from the data layer (constraint violation, conflict or
// transient failure). Its cause is only logged; the client receives a fixed message.
func IsInfrastructure(err error) bool {
	return errors.Is(err, app.ErrDataConflict) ||
		errors.Is(err, app.ErrInfraConflict) ||
		errors.Is(err, app.ErrDataViolation)
}

// buildErrorMessages formats the error messages based on status code.
// Causes of 5XX and infrastructure 409/422 never reach the body; they are logged by the caller.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
		// For 4XX BadRequest provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusNotFound:
		return nil, domainMessage(err, "not found", notFoundErrors...)
	case http.StatusConflict:
		return nil, domainMessage(err, "conflict", conflictDomainErrors...)
	case http.StatusUnprocessableEntity:
		return nil, "invalid data"
	default:
		return nil, "internal server error"
	}
}

// domainMessage returns the message of the first known domain sentinel matching err, so the body
// never carries the text of a wrapped cause. It returns fallback when no sentinel matches.
func domainMessage(err error, fallback string, sentinels ...error) string {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return s.Error()
		}
	}
	return fallback
}

func isBadRequestError(err error) bool {
	return errors.Is(err, authDomain.ErrUserPasswordDontMatch) ||
		errors.Is(err, authDomain.ErrUserPasswordTooLong) ||
		errors.Is(err, authDomain.ErrEmptyUserName) ||
		errors.Is(err, authDomain.ErrInvalidUserName) ||
		errors.Is(err, authDomain.ErrEmptyUserEmail) ||
		errors.Is(err, authDomain.ErrInvalidUserEmail) ||
		errors.Is(err, authDomain.ErrEmptyUserPassword) ||
		errors.Is(err, authDomain.ErrInvalidUserPassword) ||
		errors.Is(err, authDomain.ErrUserPasswordTooShort) ||
		errors.Is(err, jsonv2.ErrJSONSyntax) ||
		errors.Is(err, jsonv2.ErrJSONType) ||
		errors.Is(err, jsonv2.ErrJSONUnexpectedEOF) ||
		errors.Is(err, jsonv2.ErrJSONEmptyBody) ||
		errors.Is(err, jsonv2.ErrWrongPayloadFormat) ||
		errors.Is(err, vehicleDomain.ErrVehicleInvalidPlate) ||
		errors.Is(err, vehicleDomain.ErrRequiredVehicleBrand) ||
		errors.Is(err, vehicleDomain.ErrRequiredVehicleModel) ||
		errors.Is(err, vehicleDomain.ErrParamVehicleYear) ||
		errors.Is(err, vehicleDomain.ErrInvalidVehicleId) ||
		errors.Is(err, vehicleDomain.ErrInvalidSearchVehicleParams) ||
		errors.Is(err, customerDomain.ErrPhoneRequired) ||
		errors.Is(err, customerDomain.ErrCompanyNameRequired) ||
		errors.Is(err, customerDomain.ErrCompanyNameNotAllowed) ||
		errors.Is(err, customerDomain.ErrInvalidCustomerType) ||
		errors.Is(err, customerDomain.ErrInvalidDocumentFormat) ||
		errors.Is(err, customerDomain.ErrInvalidCustomerId) ||
		errors.Is(err, customerDomain.ErrInvalidCPF) ||
		errors.Is(err, customerDomain.ErrCPFLength) ||
		errors.Is(err, customerDomain.ErrInvalidCNPJ) ||
		errors.Is(err, customerDomain.ErrCNPJLength)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, vehicleDomain.ErrVehicleNotFound) ||
		errors.Is(err, customerDomain.ErrCustomerNotFound)
}

func isUnauthorizedError(err error) bool {
	return errors.Is(err, authDomain.ErrInvalidUserCredentials)
}

func isInternalServerError(err error) bool {
	return errors.Is(err, jsonv2.ErrJSONUnsupportedValue) ||
		errors.Is(err, jsonv2.ErrJSONUnsupportedType) ||
		errors.Is(err, jsonv2.ErrJSONInvalidUnmarshal)
}

func isConflictError(err error) bool {
	return errors.Is(err, app.ErrDataConflict) ||
		errors.Is(err, app.ErrInfraConflict) ||
		// TODO(F6): the delete check for service orders is not performed here anymore (orders live in
		// service-order-service); it is reintroduced via the internal REST of Service Order in F6.
		errors.Is(err, customerDomain.ErrCustomerHasServiceOrders)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, app.ErrDataViolation)
}

func isForbidden(err error) bool {
	return errors.Is(err, customerDomain.ErrInvalidCustomerProperty)
}

func unwrapAll(err error) []string {
	if err == nil {
		return nil
	}

	type multiUnwrap interface {
		Unwrap() []error
	}
	if m, ok := err.(multiUnwrap); ok {
		var all []string
		for _, e := range m.Unwrap() {
			all = append(all, unwrapAll(e)...)
		}
		return all
	}
	return []string{err.Error()}
}
