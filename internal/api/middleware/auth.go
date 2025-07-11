package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/config"
)

// BasicAuth returns a Basic Authentication middleware
func BasicAuth(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if not enabled
		if !cfg.Enabled || cfg.Username == "" || cfg.Password == "" {
			c.Next()
			return
		}

		// Get the Basic Authentication credentials
		user, pass, hasAuth := c.Request.BasicAuth()
		
		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Habibi-Go"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		// Use constant time comparison to prevent timing attacks
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(cfg.Username)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.Password)) == 1

		if !userMatch || !passMatch {
			c.Header("WWW-Authenticate", `Basic realm="Habibi-Go"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			return
		}

		c.Next()
	}
}