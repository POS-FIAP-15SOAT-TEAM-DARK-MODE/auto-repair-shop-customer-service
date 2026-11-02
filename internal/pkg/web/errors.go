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

// validationErrors are the input-validation sentinels. They answer 400, and only their own
// messages go to the errors list.
var validationErrors = []error{
	authDomain.ErrUserPasswordDontMatch,
	authDomain.ErrUserPasswordTooLong,
	authDomain.ErrEmptyUserName,
	authDomain.ErrInvalidUserName,
	authDomain.ErrEmptyUserEmail,
	authDomain.ErrInvalidUserEmail,
	authDomain.ErrEmptyUserPassword,
	authDomain.ErrInvalidUserPassword,
	authDomain.ErrUserPasswordTooShort,
	jsonv2.ErrJSONSyntax,
	jsonv2.ErrJSONType,
	jsonv2.ErrJSONUnexpectedEOF,
	jsonv2.ErrJSONEmptyBody,
	jsonv2.ErrWrongPayloadFormat,
	vehicleDomain.ErrVehicleInvalidPlate,
	vehicleDomain.ErrRequiredVehicleBrand,
	vehicleDomain.ErrRequiredVehicleModel,
	vehicleDomain.ErrParamVehicleYear,
	vehicleDomain.ErrInvalidVehicleId,
	vehicleDomain.ErrInvalidSearchVehicleParams,
	vehicleDomain.ErrVehicleNoCustomerAssociated,
	customerDomain.ErrPhoneRequired,
	customerDomain.ErrCompanyNameRequired,
	customerDomain.ErrCompanyNameNotAllowed,
	customerDomain.ErrInvalidCustomerType,
	customerDomain.ErrInvalidDocumentFormat,
	customerDomain.ErrInvalidCustomerId,
	customerDomain.ErrInvalidCPF,
	customerDomain.ErrCPFLength,
	customerDomain.ErrInvalidCNPJ,
	customerDomain.ErrCNPJLength,
}

// getHTTPStatus determines the HTTP status code for a given error.
// A joined error is a 400 only when every one of its causes is a validation sentinel; a join
// with any other cause (infrastructure or unknown) is a generic 500.
func getHTTPStatus(err error) int {
	if _, ok := validationMessages(err); ok {
		return http.StatusBadRequest
	}
	if IsJoined(err) {
		return http.StatusInternalServerError
	}
	switch {
	case isUnauthorizedError(err):
		return http.StatusUnauthorized
	case isForbidden(err):
		return http.StatusForbidden
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

// IsJoined reports whether err is a join of several errors (errors.Join or a multi-%w wrap) at its top level.
func IsJoined(err error) bool {
	_, ok := err.(multiError)
	return ok
}

type multiError interface {
	Unwrap() []error
}

// validationMessages returns the messages of the validation sentinels that make up err, and whether
// every cause of err is a validation error. A join with any other cause returns false.
func validationMessages(err error) ([]string, bool) {
	if multi, ok := err.(multiError); ok {
		var msgs []string
		for _, cause := range multi.Unwrap() {
			causeMsgs, valid := validationMessages(cause)
			if !valid {
				return nil, false
			}
			msgs = append(msgs, causeMsgs...)
		}
		return msgs, len(msgs) > 0
	}

	// A single wrap around a join: the join decides, so the wrapper text never reaches the body.
	if inner := errors.Unwrap(err); inner != nil {
		if _, ok := inner.(multiError); ok {
			return validationMessages(inner)
		}
	}

	for _, s := range validationErrors {
		if errors.Is(err, s) {
			return []string{s.Error()}, true
		}
	}
	return nil, false
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
	case http.StatusBadRequest:
		msgs, _ := validationMessages(err)
		return msgs, ""
	case http.StatusUnauthorized:
		return nil, domainMessage(err, "Unauthorized", authDomain.ErrInvalidUserCredentials)
	case http.StatusForbidden:
		return nil, domainMessage(err, "Forbidden", customerDomain.ErrInvalidCustomerProperty)
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

func isUnauthorizedError(err error) bool {
	return errors.Is(err, authDomain.ErrInvalidUserCredentials)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, vehicleDomain.ErrVehicleNotFound) ||
		errors.Is(err, customerDomain.ErrCustomerNotFound)
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
