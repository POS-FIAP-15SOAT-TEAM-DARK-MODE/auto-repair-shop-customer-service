package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		middlewareRoles  []authDomain.Role
		tokenRoles       []authDomain.Role
		tokenExpiry      time.Time
		customHeader     string
		expectedHTTPCode int
		expectedError    string
	}{
		{
			name:             "missing header",
			middlewareRoles:  []authDomain.Role{authDomain.ADMIN},
			tokenRoles:       nil,
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    "Unauthorized",
		},
		{
			name:             "malformed header",
			middlewareRoles:  []authDomain.Role{authDomain.ADMIN},
			customHeader:     "Token abc.def",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    "Unauthorized",
		},
		{
			name:             "invalid token",
			middlewareRoles:  []authDomain.Role{authDomain.ADMIN},
			customHeader:     "Bearer not-a-valid.token",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    "Unauthorized",
		},
		{
			name:             "expired token",
			middlewareRoles:  []authDomain.Role{authDomain.CUSTOMER},
			tokenRoles:       []authDomain.Role{authDomain.CUSTOMER},
			tokenExpiry:      time.Now().Add(-time.Hour),
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    "Token expired",
		},
		{
			name:             "roles mismatch",
			middlewareRoles:  []authDomain.Role{authDomain.ADMIN},
			tokenRoles:       []authDomain.Role{authDomain.CUSTOMER},
			tokenExpiry:      time.Now().Add(time.Hour),
			expectedHTTPCode: http.StatusForbidden,
			expectedError:    "Forbidden",
		},
		{
			name:             "roles match",
			middlewareRoles:  []authDomain.Role{authDomain.ATTENDANT},
			tokenRoles:       []authDomain.Role{authDomain.ATTENDANT, authDomain.ADMIN},
			tokenExpiry:      time.Now().Add(time.Hour),
			expectedHTTPCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			// protected route that returns 200 when reached
			router.GET("/protected", Auth(tt.middlewareRoles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			// set header according to test case
			if tt.customHeader != "" {
				req.Header.Set(authorizationHeader, tt.customHeader)
			} else if tt.tokenRoles != nil {
				tok, err := auth.GenerateToken("test-user", tt.tokenRoles, tt.tokenExpiry)
				if err != nil {
					t.Fatalf("failed to generate token: %v", err)
				}
				req.Header.Set(authorizationHeader, bearerPrefix+tok)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedHTTPCode, rec.Code)
			if tt.expectedHTTPCode >= http.StatusBadRequest {
				var body struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.expectedHTTPCode, body.Code)
				assert.Equal(t, tt.expectedError, body.Error)
			}
		})
	}
}
