package domain

import "errors"

var (
	ErrEmptyUserName          = errors.New("user name cannot be empty")
	ErrInvalidUserCredentials = errors.New("invalid user credentials")
	ErrInvalidUserName        = errors.New("the username must be longer than 3 characters and can only contain letters and spaces")
	ErrEmptyUserEmail         = errors.New("user email cannot be empty")
	ErrInvalidUserEmail       = errors.New("invalid user email format")
	ErrEmptyUserPassword      = errors.New("user password cannot be empty")
	ErrInvalidUserPassword    = errors.New("password must have at least 8 characters and contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
	ErrUserPasswordDontMatch  = errors.New("user passwords do not match")
	ErrUserPasswordTooLong    = errors.New("user password too long")
	ErrUserPasswordTooShort   = errors.New("user password too short")
)
