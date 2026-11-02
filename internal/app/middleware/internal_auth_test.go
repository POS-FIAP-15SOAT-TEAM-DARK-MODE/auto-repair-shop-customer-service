package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInternalRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/internal", middleware.InternalAuth())
	g.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func doInternal(t *testing.T, r *gin.Engine, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
	if token != "" {
		req.Header.Set("X-Internal-Token", token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestInternalAuth_ValidToken(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "secret-token")
	w := doInternal(t, newInternalRouter(), "secret-token")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInternalAuth_MissingOrWrongToken(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "secret-token")
	r := newInternalRouter()

	for _, token := range []string{"", "wrong", "secret-token-x"} {
		w := doInternal(t, r, token)
		require.Equal(t, http.StatusUnauthorized, w.Code, "token=%q", token)
		assert.JSONEq(t, `{"code":401,"error":"Unauthorized"}`, w.Body.String())
	}
}

func TestInternalAuth_EmptyConfiguredTokenFailsClosed(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "")
	w := doInternal(t, newInternalRouter(), "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
