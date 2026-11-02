package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "error": "Unauthorized"})
			return
		}
		c.Next()
	}
}
