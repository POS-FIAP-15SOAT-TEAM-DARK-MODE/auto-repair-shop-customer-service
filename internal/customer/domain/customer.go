package domain

import (
	"errors"
	"strings"
	"unicode"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/google/uuid"
)

type CustomerType string

const (
	IndividualCustomerType CustomerType = "INDIVIDUAL"
	CompanyCustomerType    CustomerType = "COMPANY"

	cpfLength  = 11
	cnpjLength = 14
)

func (ct CustomerType) String() string {
	return string(ct)
}

type Customer struct {
	ID          string
	UserID      string
	Type        CustomerType
	CPF         string
	CNPJ        string
	CompanyName string
	Phone       string
	User        *authDomain.User
}

// NewCustomer builds, sanitizes, and validates a Customer + embedded User.
// Password is intentionally left empty; callers must set and hash it before persisting.
func NewCustomer(name, email string, customerType CustomerType, rawDocument, companyName, phone string) (*Customer, error) {
	id := uuid.New().String()
	user := authDomain.NewUser(id, name, email, "")

	c := &Customer{
		ID:          id,
		UserID:      id,
		Type:        customerType,
		CompanyName: companyName,
		Phone:       phone,
		User:        user,
	}

	c.applyDocument(rawDocument)

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Customer) applyDocument(raw string) {
	switch c.Type {
	case IndividualCustomerType:
		c.CPF = sanitizeCPF(raw)
	case CompanyCustomerType:
		c.CNPJ = sanitizeCNPJ(raw)
	}
}

func (c *Customer) Validate() error {
	var errs []error

	if strings.TrimSpace(c.Phone) == "" {
		errs = append(errs, ErrPhoneRequired)
	}

	switch c.Type {
	case IndividualCustomerType:
		if err := validateCPF(c.CPF); err != nil {
			errs = append(errs, err)
		}
		if c.CompanyName != "" {
			errs = append(errs, ErrCompanyNameNotAllowed)
		}
	case CompanyCustomerType:
		if strings.TrimSpace(c.CompanyName) == "" {
			errs = append(errs, ErrCompanyNameRequired)
		}
		if err := validateCNPJ(c.CNPJ); err != nil {
			errs = append(errs, err)
		}
	default:
		errs = append(errs, ErrInvalidCustomerType)
	}

	if c.User != nil {
		if err := c.User.IsValidName(); err != nil {
			errs = append(errs, err)
		}
		if err := c.User.CoreUser.IsValidEmail(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// ParseDocument sanitizes a raw document, determines its type, validates it, and returns
// the canonical type and sanitized value.
func ParseDocument(raw string) (CustomerType, string, error) {
	sanitized := sanitizeCNPJ(raw) // strips non-alphanumeric and uppercases — safe for CPF too
	switch len(sanitized) {
	case cpfLength:
		if err := validateCPF(sanitized); err != nil {
			return "", "", err
		}
		return IndividualCustomerType, sanitized, nil
	case cnpjLength:
		if err := validateCNPJ(sanitized); err != nil {
			return "", "", err
		}
		return CompanyCustomerType, sanitized, nil
	default:
		return "", "", ErrInvalidDocumentFormat
	}
}

func sanitizeCPF(doc string) string {
	result := make([]rune, 0, len(doc))
	for _, r := range doc {
		if unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// sanitizeCNPJ strips formatting chars and uppercases letters (supports alphanumeric CNPJ — RFB 2243/2024).
func sanitizeCNPJ(doc string) string {
	upper := strings.ToUpper(doc)
	result := make([]rune, 0, len(upper))
	for _, r := range upper {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

func validateCPF(cpf string) error {
	if len(cpf) != cpfLength {
		return ErrCPFLength
	}

	allSame := true
	for i := 1; i < cpfLength; i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return ErrInvalidCPF
	}

	calcDigit := func(s string, length int) int {
		sum := 0
		for i := 0; i < length; i++ {
			sum += int(s[i]-'0') * (length + 1 - i)
		}
		rem := (sum * 10) % 11
		if rem == 10 {
			return 0
		}
		return rem
	}

	if calcDigit(cpf, 9) != int(cpf[9]-'0') {
		return ErrInvalidCPF
	}
	if calcDigit(cpf, 10) != int(cpf[10]-'0') {
		return ErrInvalidCPF
	}
	return nil
}

// validateCNPJ validates a Brazilian CNPJ (numeric and alphanumeric formats).
func validateCNPJ(cnpj string) error {
	if len(cnpj) != cnpjLength {
		return ErrCNPJLength
	}

	// Check digits (positions 13-14) must be numeric
	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return ErrInvalidCNPJ
	}

	for _, c := range cnpj {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return ErrInvalidCNPJ
		}
	}

	allSame := true
	for i := 1; i < cnpjLength; i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return ErrInvalidCNPJ
	}

	charValue := func(c byte) int {
		if c >= '0' && c <= '9' {
			return int(c - '0')
		}
		return int(c-'A') + 10
	}

	calcDigit := func(s string, length int, weights []int) int {
		sum := 0
		for i := 0; i < length; i++ {
			sum += charValue(s[i]) * weights[i]
		}
		rem := sum % 11
		if rem < 2 {
			return 0
		}
		return 11 - rem
	}

	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	if calcDigit(cnpj, 12, w1) != int(cnpj[12]-'0') {
		return ErrInvalidCNPJ
	}
	if calcDigit(cnpj, 13, w2) != int(cnpj[13]-'0') {
		return ErrInvalidCNPJ
	}
	return nil
}
