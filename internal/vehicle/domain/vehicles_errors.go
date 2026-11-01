package domain

import "errors"

var (
	ErrVehicleInvalidPlate         = errors.New("invalid vehicle plate")
	ErrVehicleNoCustomerAssociated = errors.New("no customer associated to the vehicle")
	ErrRequiredVehicleBrand        = errors.New("brand of the vehicle is required")
	ErrRequiredVehicleModel        = errors.New("model of the vehicle is required")
	ErrParamVehicleYear            = errors.New("year of the vehicle is not valid")
	ErrVehicleNotFound             = errors.New("vehicle not found")
	ErrInvalidVehicleId            = errors.New("invalid vehicle id")
	ErrInvalidSearchVehicleParams  = errors.New("invalid search vehicle params, a plate or customerId should be provided")
)
