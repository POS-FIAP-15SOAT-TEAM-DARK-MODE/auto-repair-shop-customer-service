package app

import "errors"

// Generic Errors
var (
	ErrDataConflict  = errors.New("resource already exists")
	ErrDataViolation = errors.New("data violates business constraints")
	ErrInfraConflict = errors.New("something went wrong with our resources, try again")
	// ErrUpstreamUnavailable is returned when a call to another service fails (network, timeout, 5XX).
	ErrUpstreamUnavailable = errors.New("upstream unavailable")
)
