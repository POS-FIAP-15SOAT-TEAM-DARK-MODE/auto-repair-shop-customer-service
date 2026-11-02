package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
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
			// The submitted token is never logged, only the route that rejected it.
			logger.Of(c).Warn("internal.auth_failed", zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "error": "Unauthorized"})
			return
		}
		c.Next()
	}
}
