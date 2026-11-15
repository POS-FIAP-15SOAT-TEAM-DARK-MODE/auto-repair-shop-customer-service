package domain

import "errors"

var (
	ErrPhoneRequired            = errors.New("phone is required")
	ErrCompanyNameRequired      = errors.New("companyName is required for COMPANY type")
	ErrCompanyNameNotAllowed    = errors.New("companyName is not allowed for INDIVIDUAL type")
	ErrInvalidCustomerType      = errors.New("invalid customer type")
	ErrInvalidCPF               = errors.New("invalid CPF")
	ErrInvalidCNPJ              = errors.New("invalid CNPJ")
	ErrCPFLength                = errors.New("CPF must have 11 digits")
	ErrCNPJLength               = errors.New("CNPJ must have 14 characters")
	ErrCustomerNotFound         = errors.New("customer not found")
	ErrCustomerHasServiceOrders = errors.New("customer has associated service orders")
	ErrInvalidDocumentFormat    = errors.New("document must be a valid CPF (11 digits) or CNPJ (14 characters)")
	ErrInvalidCustomerId        = errors.New("invalid customer id")
	ErrInvalidCustomerProperty  = errors.New("customer not allowed to execute this operation")
)
