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

// buildErrorMessages formats the error messages based on status code.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
		// For 4XX BadRequest provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusNotFound:
		return nil, err.Error()
	case http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		// For 409/422 and 5XX unwrap recursively and return the innermost (root cause) error message.
		root := err
		for {
			unwrapped := errors.Unwrap(root)
			if unwrapped == nil {
				break
			}
			root = unwrapped
		}
		return nil, root.Error()
	default:
		// For any other case, generic error message.
		return nil, "Internal Server Error"
	}
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
		errors.Is(err, customerDomain.ErrInvalidCustomerId)
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
