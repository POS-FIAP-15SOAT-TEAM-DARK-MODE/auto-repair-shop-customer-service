package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/interfaces/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		mockSvc        func(t *testing.T) *mocks.AuthService
		expectedStatus int
	}{
		{
			name: "Should fail gracefully when request body is invalid",
			requestBody: `{
				"email": "test@example.com"
				"password": "password"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail gracefully when service fails",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusInternalServerError,
			mockSvc: func(t *testing.T) *mocks.AuthService {
				svc := mocks.NewAuthService(t)
				svc.On("Login", mock.Anything, mock.Anything).Return(adapters.LoginResponse{}, errors.New("service failed"))
				return svc
			},
		},
		{
			name:           "Should login successfully",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusOK,
			mockSvc: func(t *testing.T) *mocks.AuthService {
				svc := mocks.NewAuthService(t)
				svc.On("Login", mock.Anything, mock.Anything).
					Return(adapters.LoginResponse{}, nil)
				return svc
			},
		},
	}

	for _, tt := range tests {
		loginRoute := "/v1/auth/login"
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.AuthService{}
			if tt.mockSvc != nil {
				mockSvc = tt.mockSvc(t)
			}

			h := auth.NewController(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodPost, loginRoute, bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.Login)(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		body       string
		mockSvc    func(t *testing.T) *mocks.AuthService
		expectedSC int
	}{
		{
			name:       "Should fail gracefully when request body is invalid",
			path:       "123",
			body:       `{"role":`,
			expectedSC: http.StatusBadRequest,
		},
		{
			name: "Should fail gracefully when service fails",
			path: "123",
			body: `{"role":"MECHANIC"}`,
			mockSvc: func(t *testing.T) *mocks.AuthService {
				svc := mocks.NewAuthService(t)
				svc.EXPECT().ChangeRole(mock.Anything, "123", mock.AnythingOfType("adapters.ChangeRoleRequest")).Return(errors.New("service failed"))
				return svc
			},
			expectedSC: http.StatusInternalServerError,
		},
		{
			name: "Should update role successfully",
			path: "123",
			body: `{"role":"MECHANIC"}`,
			mockSvc: func(t *testing.T) *mocks.AuthService {
				svc := mocks.NewAuthService(t)
				svc.EXPECT().ChangeRole(mock.Anything, "123", mock.AnythingOfType("adapters.ChangeRoleRequest")).Return(nil)
				return svc
			},
			expectedSC: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.AuthService{}
			if tt.mockSvc != nil {
				mockSvc = tt.mockSvc(t)
			}

			h := auth.NewController(mockSvc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.path}}
			url := fmt.Sprintf("/v1/users/%s/role", tt.path)
			req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinOnlyErrorHandler(h.ChangeRole)(c)
			assert.Equal(t, tt.expectedSC, w.Code)
		})
	}
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := mocks.NewAuthService(t)
	h := auth.NewController(svc)

	svc.EXPECT().Register(mock.Anything, mock.AnythingOfType("adapters.CreateUserRequest")).Return(adapters.UserResponse{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := map[string]string{
		"name":            "Test User",
		"email":           "test@example.com",
		"password":        "Secret@123",
		"confirmPassword": "Secret@123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.Register)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := auth.NewController(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.Register)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := mocks.NewAuthService(t)
	h := auth.NewController(svc)

	svc.EXPECT().Register(mock.Anything, mock.AnythingOfType("adapters.CreateUserRequest")).Return(adapters.UserResponse{}, errors.New("internal error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := map[string]string{
		"name":            "Test User",
		"email":           "test@example.com",
		"password":        "Secret@123",
		"confirmPassword": "Secret@123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.Register)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
