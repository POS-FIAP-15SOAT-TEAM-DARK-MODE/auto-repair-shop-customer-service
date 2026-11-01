package customer_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
	adapters2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/customer/interfaces/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func validCustomerResponse(tp domain.CustomerType) adapters.CustomerResponse {
	return adapters.CustomerResponse{
		ID:       "uuid-individual",
		Type:     tp.String(),
		Document: "11144477735",
		Phone:    "11999999999",
		UserResponse: adapters2.UserResponse{
			Name:  "João Silva",
			Email: "joao@example.com",
		},
	}
}

// --- Create ---

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]any
		mockSetup      func(*mocks.CustomerService)
		expectedStatus int
	}{
		{
			name: "success_individual",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("adapters.CreateCustomerRequest")).Return(validCustomerResponse(domain.IndividualCustomerType), nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "success_company",
			body: map[string]any{
				"name": "Empresa SA", "email": "empresa@example.com",
				"password": "Senha@123", "type": "COMPANY",
				"document": "11.222.333/0001-81", "companyName": "Empresa SA", "phone": "1133334444",
			},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("adapters.CreateCustomerRequest")).Return(validCustomerResponse(domain.CompanyCustomerType), nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid_json",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service_conflict",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("adapters.CreateCustomerRequest")).Return(adapters.CustomerResponse{}, app.ErrDataConflict)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service_error_returns_500",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("adapters.CreateCustomerRequest")).Return(adapters.CustomerResponse{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewCustomerService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := customer.NewController(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var reqBody *bytes.Reader
			if tt.body != nil {
				payload, _ := json.Marshal(tt.body)
				reqBody = bytes.NewReader(payload)
			} else {
				reqBody = bytes.NewReader([]byte("invalid-json{"))
			}

			req := httptest.NewRequest(http.MethodPost, "/customers", reqBody)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			routing.GinHandler(h.Create, http.StatusCreated)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// --- GetByDocument ---

func TestGetByDocument(t *testing.T) {
	tests := []struct {
		name           string
		queryParam     string
		mockSetup      func(*mocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name:       "success",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, "111.444.777-35").Return(validCustomerResponse(domain.IndividualCustomerType), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name:       "not_found_returns_404",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(adapters.CustomerResponse{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
			assertBody:     nil,
		},
		{
			name:       "invalid_document_format_returns_400",
			queryParam: "document=123",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(adapters.CustomerResponse{}, domain.ErrInvalidDocumentFormat)
			},
			expectedStatus: http.StatusBadRequest,
			assertBody:     nil,
		},
		{
			name:       "service_error_returns_500",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(adapters.CustomerResponse{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewCustomerService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := customer.NewController(svc)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			url := "/customers"
			if tt.queryParam != "" {
				url += "?" + tt.queryParam
			}

			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			routing.GinHandler(h.GetByDocument)(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.assertBody != nil {
				var body map[string]any
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				tt.assertBody(t, body)
			}
		})
	}
}

// --- GetByID ---

func TestGetByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*mocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name: "success",
			id:   "uuid-individual",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validCustomerResponse(domain.IndividualCustomerType), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name: "not_found_returns_404",
			id:   "unknown",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "unknown").Return(adapters.CustomerResponse{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
			assertBody:     nil,
		},
		{
			name: "service_error_returns_500",
			id:   "any-id",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "any-id").Return(adapters.CustomerResponse{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewCustomerService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := customer.NewController(svc)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			req := httptest.NewRequest(http.MethodGet, "/customers/"+tt.id, nil)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			routing.GinHandler(h.GetByID)(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.assertBody != nil {
				var body map[string]any
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				tt.assertBody(t, body)
			}
		})
	}
}

// --- Update ---

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           map[string]any
		mockSetup      func(*mocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name: "success_all_fields",
			id:   "uuid-individual",
			body: map[string]any{"name": "Novo Nome", "email": "novo@example.com", "phone": "11888888888"},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", mock.AnythingOfType("adapters.UpdateCustomerRequest")).
					Return(validCustomerResponse(domain.IndividualCustomerType), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name: "success_partial_only_phone",
			id:   "uuid-individual",
			body: map[string]any{"phone": "11777777777"},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", mock.AnythingOfType("adapters.UpdateCustomerRequest")).
					Return(validCustomerResponse(domain.CompanyCustomerType), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not_found_returns_404",
			id:   "unknown",
			body: map[string]any{"name": "Novo Nome"},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "unknown", mock.AnythingOfType("adapters.UpdateCustomerRequest")).
					Return(adapters.CustomerResponse{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid_json_returns_400",
			id:             "uuid-individual",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate_email_returns_409",
			id:   "uuid-individual",
			body: map[string]any{"email": "duplicate@example.com"},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", mock.AnythingOfType("adapters.UpdateCustomerRequest")).
					Return(adapters.CustomerResponse{}, app.ErrDataConflict)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service_error_returns_500",
			id:   "uuid-individual",
			body: map[string]any{"phone": "11888888888"},
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", mock.AnythingOfType("adapters.UpdateCustomerRequest")).
					Return(adapters.CustomerResponse{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewCustomerService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := customer.NewController(svc)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}

			var reqBody *bytes.Reader
			if tt.body != nil {
				payload, _ := json.Marshal(tt.body)
				reqBody = bytes.NewReader(payload)
			} else {
				reqBody = bytes.NewReader([]byte("invalid-json{"))
			}

			req := httptest.NewRequest(http.MethodPut, "/customers/"+tt.id, reqBody)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			routing.GinHandler(h.Update)(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.assertBody != nil {
				var body map[string]any
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				tt.assertBody(t, body)
			}
		})
	}
}

// --- Delete ---

func TestDelete(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*mocks.CustomerService)
		expectedStatus int
	}{
		{
			name: "success_returns_204",
			id:   "uuid-individual",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "not_found_returns_404",
			id:   "unknown",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "unknown").Return(domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "has_service_orders_returns_409",
			id:   "uuid-individual",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(domain.ErrCustomerHasServiceOrders)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service_error_returns_500",
			id:   "uuid-individual",
			mockSetup: func(svc *mocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewCustomerService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := customer.NewController(svc)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			req := httptest.NewRequest(http.MethodDelete, "/customers/"+tt.id, nil)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			routing.GinOnlyErrorHandler(h.Delete)(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
