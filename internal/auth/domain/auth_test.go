package domain_test

import (
	"context"
	"strings"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_SetsFields(t *testing.T) {
	u := domain.NewUser("id-1", "João Silva", "joao@email.com", "Secret!1")
	assert.Equal(t, "id-1", u.ID)
	assert.Equal(t, "João Silva", u.Name)
	assert.Equal(t, "joao@email.com", u.Email)
	assert.Equal(t, "Secret!1", u.Password)
}

func TestIsValidEmail(t *testing.T) {
	cases := map[string]error{
		"joao@email.com":  nil,
		"":                domain.ErrEmptyUserEmail,
		"joao @email.com": domain.ErrInvalidUserEmail,
		"joao.email.com":  domain.ErrInvalidUserEmail,
		"a@b@c.com":       domain.ErrInvalidUserEmail,
		"@email.com":      domain.ErrInvalidUserEmail,
		"joao@emailcom":   domain.ErrInvalidUserEmail,
		"joao@.email.com": domain.ErrInvalidUserEmail,
		"joao@email.com.": domain.ErrInvalidUserEmail,
	}
	for email, want := range cases {
		u := domain.CoreUser{Email: email}
		assert.ErrorIs(t, u.IsValidEmail(), want, "email=%q", email)
	}
}

func TestIsValidPassword(t *testing.T) {
	cases := map[string]error{
		"Secret!1":                nil,
		"":                        domain.ErrEmptyUserPassword,
		"Sh!1":                    domain.ErrUserPasswordTooShort,
		"secret!1":                domain.ErrInvalidUserPassword,
		"Secretsecret":            domain.ErrInvalidUserPassword,
		strings.Repeat("Aa!", 30): domain.ErrUserPasswordTooLong,
	}
	for pw, want := range cases {
		c := domain.CoreUser{Password: pw}
		assert.ErrorIs(t, c.IsValidPassword(), want, "password=%q", pw)
	}
}

func TestCheckRequiredFields_JoinsErrors(t *testing.T) {
	c := domain.CoreUser{}
	err := c.CheckRequiredFields()
	assert.ErrorIs(t, err, domain.ErrEmptyUserPassword)
	assert.ErrorIs(t, err, domain.ErrEmptyUserEmail)

	ok := domain.CoreUser{Email: "a@b.co", Password: "x"}
	assert.NoError(t, ok.CheckRequiredFields())
}

func TestCoreUserValidate(t *testing.T) {
	assert.NoError(t, (&domain.CoreUser{Email: "a@b.co", Password: "Secret!1"}).Validate())

	err := (&domain.CoreUser{Email: "bad", Password: "weak"}).Validate()
	assert.ErrorIs(t, err, domain.ErrInvalidUserEmail)
	assert.ErrorIs(t, err, domain.ErrUserPasswordTooShort)
}

func TestHashAndComparePassword(t *testing.T) {
	c := &domain.CoreUser{Password: "Secret!1"}
	require.NoError(t, c.HashPassword(context.Background()))
	assert.NotEqual(t, "Secret!1", c.Password)

	assert.NoError(t, c.ComparePassword("Secret!1"))
	assert.ErrorIs(t, c.ComparePassword("wrong"), domain.ErrInvalidUserCredentials)
}

func TestHashPassword_TooLongForBcrypt(t *testing.T) {
	c := &domain.CoreUser{Password: strings.Repeat("a", 100)}
	assert.ErrorIs(t, c.HashPassword(context.Background()), domain.ErrUserPasswordTooLong)
}

func TestUserValidateAndName(t *testing.T) {
	valid := &domain.User{ID: "1", Name: "João Silva", CoreUser: domain.CoreUser{Email: "j@s.com", Password: "Secret!1"}}
	assert.NoError(t, valid.Validate())

	assert.ErrorIs(t, (&domain.User{Name: "   "}).IsValidName(), domain.ErrEmptyUserName)
	assert.ErrorIs(t, (&domain.User{Name: "Jo"}).IsValidName(), domain.ErrInvalidUserName)
	assert.ErrorIs(t, (&domain.User{Name: "João123"}).IsValidName(), domain.ErrInvalidUserName)

	bad := &domain.User{Name: "Jo", CoreUser: domain.CoreUser{Email: "bad", Password: "weak"}}
	err := bad.Validate()
	assert.ErrorIs(t, err, domain.ErrInvalidUserName)
	assert.ErrorIs(t, err, domain.ErrInvalidUserEmail)
}

func TestRoleGroups(t *testing.T) {
	assert.Contains(t, domain.AttendantRoles, domain.ATTENDANT)
	assert.Contains(t, domain.MechanicRoles, domain.MECHANIC)
	assert.Contains(t, domain.CustomerRoles, domain.CUSTOMER)
	assert.Contains(t, domain.AttendantAndMechanicRoles, domain.MECHANIC)
	assert.Len(t, domain.ValidRoles, 4)
}
