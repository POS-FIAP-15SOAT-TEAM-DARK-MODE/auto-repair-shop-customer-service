package domain

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type (
	CoreUser struct {
		Email    string
		Password string
	}

	User struct {
		CoreUser
		ID   string
		Name string
	}

	Role string
)

func NewUser(id, name, email, password string) *User {
	return &User{
		ID:   id,
		Name: name,
		CoreUser: CoreUser{
			Email:    email,
			Password: password,
		},
	}
}

const (
	ADMIN     Role = "ADMIN"
	ATTENDANT Role = "ATTENDANT"
	MECHANIC  Role = "MECHANIC"
	CUSTOMER  Role = "CUSTOMER"
)

var (
	AttendantRoles            = []Role{ATTENDANT, ADMIN}
	MechanicRoles             = []Role{MECHANIC, ADMIN}
	CustomerRoles             = []Role{CUSTOMER, ADMIN}
	AttendantAndMechanicRoles = []Role{ATTENDANT, MECHANIC, ADMIN}
	ValidRoles                = map[Role]struct{}{
		ADMIN:     struct{}{},
		ATTENDANT: struct{}{},
		MECHANIC:  struct{}{},
		CUSTOMER:  struct{}{},
	}
)

var (
	passwordMinLength  = 8
	passwordMaxLength  = 80
	minNameLength      = 3
	passwordHasUpper   = regexp.MustCompile(`[A-Z]`)
	passwordHasSpecial = regexp.MustCompile(`[^a-zA-Z0-9]`)
	nameIsValid        = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s]+$`)
)

func (c *CoreUser) CheckRequiredFields() error {
	var errs []error
	if err := c.isPasswordEmpty(); err != nil {
		errs = append(errs, err)
	}

	if err := c.isEmailEmpty(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (c *CoreUser) Validate() error {
	var errs []error
	if err := c.IsValidEmail(); err != nil {
		errs = append(errs, err)
	}

	if err := c.IsValidPassword(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (c *CoreUser) IsValidEmail() error {
	email := strings.TrimSpace(c.Email)
	if err := c.isEmailEmpty(); err != nil {
		return err
	}

	if strings.Contains(email, " ") {
		return ErrInvalidUserEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrInvalidUserEmail
	}

	localPart := parts[0]
	domainPart := parts[1]

	if localPart == "" || domainPart == "" {
		return ErrInvalidUserEmail
	}

	if !strings.Contains(domainPart, ".") {
		return ErrInvalidUserEmail
	}

	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return ErrInvalidUserEmail
	}

	return nil
}

func (c *CoreUser) isEmailEmpty() error {
	if c.Email == "" {
		return ErrEmptyUserEmail
	}
	return nil
}

func (c *CoreUser) IsValidPassword() error {
	if err := c.isPasswordEmpty(); err != nil {
		return err
	}

	if len(c.Password) > passwordMaxLength {
		return ErrUserPasswordTooLong
	}
	if len(c.Password) < passwordMinLength {
		return ErrUserPasswordTooShort
	}
	if !passwordHasUpper.MatchString(c.Password) || !passwordHasSpecial.MatchString(c.Password) {
		return ErrInvalidUserPassword
	}

	return nil
}

func (c *CoreUser) isPasswordEmpty() error {
	if c.Password == "" {
		return ErrEmptyUserPassword
	}
	return nil
}

var hashingCost int

func init() {
	hashingCost = env.GetInt("BCRYPT_COST", bcrypt.DefaultCost)
}

func (c *CoreUser) HashPassword(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), hashingCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return ErrUserPasswordTooLong
		}

		logger.Of(ctx).Warn("user.hash_failed", zap.String("email", c.Email), zap.Error(err))
		return err
	}
	c.Password = string(hash)
	return nil
}

func (c *CoreUser) ComparePassword(value string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(value)); err != nil {
		return ErrInvalidUserCredentials
	}
	return nil
}

func (u *User) Validate() error {
	var errs []error
	if err := u.IsValidName(); err != nil {
		errs = append(errs, err)
	}
	if err := u.CoreUser.Validate(); err != nil {
		if uw, ok := err.(interface{ Unwrap() []error }); ok {
			errSlice := uw.Unwrap()
			errs = append(errs, errSlice...)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (u *User) IsValidName() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrEmptyUserName
	}
	if utf8.RuneCountInString(u.Name) < minNameLength {
		return ErrInvalidUserName
	}

	if !nameIsValid.MatchString(u.Name) {
		return ErrInvalidUserName
	}

	return nil
}
