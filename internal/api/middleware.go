package api

import (
	"net/http"
	"strings"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gin-gonic/gin"
)

// TODO: Not the best auth middleware, but good enough for testing.
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := extractToken(ctx)
		if token != config.Get().Server.Token {
			ctx.String(http.StatusUnauthorized, "Unauthorized")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func extractToken(ctx *gin.Context) string {
	bearerToken := ctx.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
