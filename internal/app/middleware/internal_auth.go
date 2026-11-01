package middleware

import (
	"crypto/subtle"
	"os"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

// InternalAuth protects the /internal/* routes, which are only called between services.
// It compares the X-Internal-Token header with INTERNAL_TOKEN in constant time.
// An empty INTERNAL_TOKEN rejects every request (fail closed).
func InternalAuth() gin.HandlerFunc {
	expected := os.Getenv("INTERNAL_TOKEN")
	return func(c *gin.Context) {
		got := c.GetHeader(internalTokenHeader)
		if expected == "" || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
			status, body := web.Error(domain.ErrInvalidUserCredentials)
			c.AbortWithStatusJSON(status, body)
			return
		}
		c.Next()
	}
}
