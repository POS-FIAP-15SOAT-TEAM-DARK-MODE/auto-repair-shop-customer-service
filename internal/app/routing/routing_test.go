package routing_test

import (
	"context"
	"encoding/json"
	goHttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
	authAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	authMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces/mocks"
	customerAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	customerMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
	vehicleAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/adapters"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/domain"
	vehiclesMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/vehicle/interfaces/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const internalToken = "routing-test-internal-token"

// fakeCustomerInternal implements CustomerInternalHTTPController with canned results.
type fakeCustomerInternal struct {
	resp customerAdapters.InternalCustomerResponse
	err  error
}

func (f fakeCustomerInternal) GetInternalByID(_ context.Context, _ *goHttp.Request) (customerAdapters.InternalCustomerResponse, error) {
	return f.resp, f.err
}

// fakeVehicleInternal implements VehicleInternalHTTPController with canned results.
type fakeVehicleInternal struct {
	resp vehicleAdapters.InternalVehicleResponse
	err  error
}

func (f fakeVehicleInternal) GetInternalByID(_ context.Context, _ *goHttp.Request) (vehicleAdapters.InternalVehicleResponse, error) {
	return f.resp, f.err
}

func newRouter(t *testing.T, ci fakeCustomerInternal, vi fakeVehicleInternal) (*gin.Engine, *customerMocks.CustomerHTTPController, *vehiclesMocks.VehicleHTTPController, *authMocks.AuthHTTPController) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("INTERNAL_TOKEN", internalToken)

	mockCustomer := customerMocks.NewCustomerHTTPController(t)
	mockVehicle := vehiclesMocks.NewVehicleHTTPController(t)
	mockAuth := authMocks.NewAuthHTTPController(t)

	c := &container.HandlersWrapper{
		UserHandler:             mockAuth,
		CustomerHandler:         mockCustomer,
		VehicleHandler:          mockVehicle,
		CustomerInternalHandler: ci,
		VehicleInternalHandler:  vi,
	}
	m := &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"ErrorHandler": middleware.ErrorHandler(),
	}
	return routing.SetupRouter(c, m), mockCustomer, mockVehicle, mockAuth
}

func do(r *gin.Engine, method, path, token string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func tokenFor(t *testing.T, role authDomain.Role) string {
	t.Helper()
	tok, err := auth.GenerateToken("user-id", []authDomain.Role{role}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	return tok
}

func TestSetupRouter_PublicAndPing(t *testing.T) {
	r, _, _, mockAuth := newRouter(t, fakeCustomerInternal{}, fakeVehicleInternal{})
	mockAuth.EXPECT().Login(mock.Anything, mock.AnythingOfType("*http.Request")).Return(authAdapters.LoginResponse{}, nil)

	assert.Equal(t, goHttp.StatusOK, do(r, goHttp.MethodGet, "/ping", "", "").Code)
	assert.Equal(t, goHttp.StatusOK, do(r, goHttp.MethodPost, "/v1/auth/login", "", "{}").Code)
	assert.Equal(t, goHttp.StatusOK, do(r, goHttp.MethodGet, "/swagger.yaml", "", "").Code)
}

func TestSetupRouter_AuthRoutes(t *testing.T) {
	r, _, _, mockAuth := newRouter(t, fakeCustomerInternal{}, fakeVehicleInternal{})
	mockAuth.EXPECT().Register(mock.Anything, mock.AnythingOfType("*http.Request")).Return(authAdapters.UserResponse{}, nil)
	mockAuth.EXPECT().ChangeRole(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	assert.Equal(t, goHttp.StatusCreated, do(r, goHttp.MethodPost, "/v1/auth/register", tokenFor(t, authDomain.ADMIN), "{}").Code)
	assert.Equal(t, goHttp.StatusForbidden, do(r, goHttp.MethodPost, "/v1/auth/register", tokenFor(t, authDomain.ATTENDANT), "{}").Code)
	assert.Equal(t, goHttp.StatusNoContent, do(r, goHttp.MethodPatch, "/v1/users/u1/role", tokenFor(t, authDomain.ADMIN), "{}").Code)
}

func TestSetupRouter_CustomerRoutes(t *testing.T) {
	r, mockCustomer, _, _ := newRouter(t, fakeCustomerInternal{}, fakeVehicleInternal{})
	mockCustomer.EXPECT().GetByID(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	assert.Equal(t, goHttp.StatusOK, do(r, goHttp.MethodGet, "/v1/customers/c1", tokenFor(t, authDomain.ATTENDANT), "").Code)
	assert.Equal(t, goHttp.StatusCreated, do(r, goHttp.MethodPost, "/v1/customers", tokenFor(t, authDomain.ATTENDANT), "{}").Code)
	assert.Equal(t, goHttp.StatusNoContent, do(r, goHttp.MethodDelete, "/v1/customers/c1", tokenFor(t, authDomain.ATTENDANT), "").Code)
	assert.Equal(t, goHttp.StatusForbidden, do(r, goHttp.MethodGet, "/v1/customers/c1", tokenFor(t, authDomain.MECHANIC), "").Code)
	assert.Equal(t, goHttp.StatusUnauthorized, do(r, goHttp.MethodGet, "/v1/customers/c1", "", "").Code)
}

func TestSetupRouter_VehicleRoutes(t *testing.T) {
	r, _, mockVehicle, _ := newRouter(t, fakeCustomerInternal{}, fakeVehicleInternal{})
	mockVehicle.EXPECT().List(mock.Anything, mock.AnythingOfType("*http.Request")).Return(vehicleAdapters.PaginatedVehicleResponse{}, nil)
	mockVehicle.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(vehicleAdapters.VehicleResponse{}, nil)

	assert.Equal(t, goHttp.StatusOK, do(r, goHttp.MethodGet, "/v1/vehicles", tokenFor(t, authDomain.MECHANIC), "").Code)
	assert.Equal(t, goHttp.StatusCreated, do(r, goHttp.MethodPost, "/v1/vehicles", tokenFor(t, authDomain.ATTENDANT), "{}").Code)
	assert.Equal(t, goHttp.StatusForbidden, do(r, goHttp.MethodPost, "/v1/vehicles", tokenFor(t, authDomain.MECHANIC), "{}").Code)
}

func TestSetupRouter_RemovedDomainsAreNotRouted(t *testing.T) {
	r, _, _, _ := newRouter(t, fakeCustomerInternal{}, fakeVehicleInternal{})
	for _, p := range []string{"/v1/works", "/v1/supplies", "/v1/service-order", "/v1/reports/average-execution-time"} {
		assert.Equal(t, goHttp.StatusNotFound, do(r, goHttp.MethodGet, p, tokenFor(t, authDomain.ATTENDANT), "").Code, p)
	}
}

func TestSetupRouter_InternalRoutes(t *testing.T) {
	ci := fakeCustomerInternal{resp: customerAdapters.InternalCustomerResponse{
		ID: "c1", UserID: "u1", Type: "INDIVIDUAL", Name: "João Silva", Email: "joao.silva@email.com",
		Document: "52998224725", CompanyName: nil, Phone: "11999990001",
	}}
	vi := fakeVehicleInternal{resp: vehicleAdapters.InternalVehicleResponse{
		ID: "v1", CustomerID: "c1", LicensePlate: "ABC1D23", Brand: "Fiat", Model: "Uno", Year: 2018,
	}}
	r, _, _, _ := newRouter(t, ci, vi)

	w := do(r, goHttp.MethodGet, "/internal/customers/c1", "", "")
	// no token => 401 (route exists, guarded)
	assert.Equal(t, goHttp.StatusUnauthorized, w.Code)

	req := httptest.NewRequest(goHttp.MethodGet, "/internal/customers/c1", nil)
	req.Header.Set("X-Internal-Token", internalToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, goHttp.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "c1", body["id"])
	assert.Equal(t, "u1", body["user_id"])
	assert.Equal(t, "52998224725", body["document"])
	assert.Equal(t, "joao.silva@email.com", body["email"])
	assert.Equal(t, "11999990001", body["phone"])
	assert.Nil(t, body["company_name"])
	assert.Len(t, body, 8)

	req = httptest.NewRequest(goHttp.MethodGet, "/internal/vehicles/v1", nil)
	req.Header.Set("X-Internal-Token", internalToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, goHttp.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"license_plate":"ABC1D23"`)
	assert.Contains(t, w.Body.String(), `"customer_id":"c1"`)
}

func TestSetupRouter_InternalNotFound(t *testing.T) {
	ci := fakeCustomerInternal{err: customerDomain.ErrCustomerNotFound}
	r, _, _, _ := newRouter(t, ci, fakeVehicleInternal{err: vehicleDomain.ErrVehicleNotFound})

	for _, p := range []string{"/internal/customers/missing", "/internal/vehicles/missing"} {
		req := httptest.NewRequest(goHttp.MethodGet, p, nil)
		req.Header.Set("X-Internal-Token", internalToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, goHttp.StatusNotFound, w.Code, p)
	}
}
