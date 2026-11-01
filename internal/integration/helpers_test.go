//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
	"github.com/stretchr/testify/require"
)

// ─── Response types ───────────────────────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type customerResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Document    string `json:"document"`
	CompanyName string `json:"companyName,omitempty"`
	Phone       string `json:"phone"`
}

type vehicleResponse struct {
	ID           string `json:"id"`
	LicensePlate string `json:"licensePlate"`
	BrandModel   string `json:"brandModel"`
	Year         int    `json:"year"`
	CustomerID   string `json:"customerId"`
}

type paginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int64 `json:"totalPages"`
	PageSize   int64 `json:"pageSize"`
	Page       int64 `json:"page"`
}

// ─── Token helpers ────────────────────────────────────────────────────────────

func adminToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateToken("admin-test-user", []authDomain.Role{authDomain.ADMIN}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

func attendantToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateToken("attendant-test-user", []authDomain.Role{authDomain.ATTENDANT}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

func mechanicToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateToken("mechanic-test-user", []authDomain.Role{authDomain.MECHANIC}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

func customerToken(t *testing.T, userID string) string {
	t.Helper()
	tok, err := auth.GenerateToken(userID, []authDomain.Role{authDomain.CUSTOMER}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

// ─── HTTP client helpers ──────────────────────────────────────────────────────

func doRequest(t *testing.T, method, path string, body any, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, testServer.URL+path, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var out T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return b
}

// ─── Fixture helpers ──────────────────────────────────────────────────────────

func createCustomer(t *testing.T, name, email, cpf, phone string) customerResponse {
	t.Helper()
	body := map[string]any{
		"name":     name,
		"email":    email,
		"password": "Test@12345",
		"type":     "INDIVIDUAL",
		"document": cpf,
		"phone":    phone,
	}
	resp := doRequest(t, http.MethodPost, "/v1/customers", body, attendantToken(t))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createCustomer: unexpected status %d body=%s", resp.StatusCode, readBodyStr(t, resp))
	}
	return decodeJSON[customerResponse](t, resp)
}

func createCompanyCustomer(t *testing.T, name, email, cnpj, companyName, phone string) customerResponse {
	t.Helper()
	body := map[string]any{
		"name":        name,
		"email":       email,
		"password":    "Test@12345",
		"type":        "COMPANY",
		"document":    cnpj,
		"companyName": companyName,
		"phone":       phone,
	}
	resp := doRequest(t, http.MethodPost, "/v1/customers", body, attendantToken(t))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createCompanyCustomer: unexpected status %d body=%s", resp.StatusCode, readBodyStr(t, resp))
	}
	return decodeJSON[customerResponse](t, resp)
}

func createVehicle(t *testing.T, plate, brand, model string, year int, customerID string) vehicleResponse {
	t.Helper()
	body := map[string]any{
		"licensePlate": plate,
		"brand":        brand,
		"model":        model,
		"year":         year,
		"customerId":   customerID,
	}
	resp := doRequest(t, http.MethodPost, "/v1/vehicles", body, attendantToken(t))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createVehicle: unexpected status %d body=%s", resp.StatusCode, readBodyStr(t, resp))
	}
	return decodeJSON[vehicleResponse](t, resp)
}

func readBodyStr(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return string(b)
}

// ─── Unique test data generators ─────────────────────────────────────────────

var (
	cpfCounter   int64
	emailCounter int64
	plateCounter int64
	phoneCounter int64 = 11900000000
)

// nextCPF generates a unique, algorithmically valid CPF for each call.
func nextCPF() string {
	cpfCounter++
	base := cpfCounter + 100_000_000

	d := [11]int{}
	d[0] = int(base/100000000) % 10
	d[1] = int(base/10000000) % 10
	d[2] = int(base/1000000) % 10
	d[3] = int(base/100000) % 10
	d[4] = int(base/10000) % 10
	d[5] = int(base/1000) % 10
	d[6] = int(base/100) % 10
	d[7] = int(base/10) % 10
	d[8] = int(base) % 10

	sum := 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (10 - i)
	}
	rem := sum % 11
	if rem < 2 {
		d[9] = 0
	} else {
		d[9] = 11 - rem
	}

	sum = 0
	for i := 0; i < 10; i++ {
		sum += d[i] * (11 - i)
	}
	rem = sum % 11
	if rem < 2 {
		d[10] = 0
	} else {
		d[10] = 11 - rem
	}

	return fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d",
		d[0], d[1], d[2], d[3], d[4], d[5], d[6], d[7], d[8], d[9], d[10])
}

func nextEmail() string {
	emailCounter++
	return fmt.Sprintf("test%d@integration.test", emailCounter)
}

// nextPlate generates a unique license plate in old format (ITG-NNNN).
func nextPlate() string {
	plateCounter++
	return fmt.Sprintf("ITG-%04d", plateCounter)
}

func nextPhone() string {
	phoneCounter++
	return fmt.Sprintf("%d", phoneCounter)
}

// Known valid CNPJ for company customer tests.
const cnpj1 = "45.997.418/0001-53"

// cnpj1Sanitized is cnpj1 with all non-alphanumeric chars stripped (what the API stores/returns).
const cnpj1Sanitized = "45997418000153"
