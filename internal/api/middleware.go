package api

import (
	"net/http"
	"strings"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gin-gonic/gin"
)

// TODO: Not the best auth middleware, but good enough for testing.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != config.Get().Server.Token {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
