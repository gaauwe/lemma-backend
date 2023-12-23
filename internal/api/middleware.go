package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/lemmy"
	"github.com/gaauwe/lemma-backend/internal/lemmy/types"
	"github.com/gin-gonic/gin"
)

// TODO: Not the best auth middleware, but good enough for testing.
func AdminAuthMiddleware() gin.HandlerFunc {
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

func UserAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username := ctx.Param("username")
		token := extractToken(ctx)
		user, err := database.GetUserByUsername(username)

		// Check if user exists.
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Unauthorized")
			ctx.Abort()
			return
		}

		// Check if token is valid for user.
		server := fmt.Sprintf("https://%s", strings.Split(user.Username, "@")[1])
		c, err := lemmy.New(server)
		c.Token = token
		site, err := c.Site(ctx, types.GetSite{Auth: types.NewOptional(c.Token)})

		if err != nil || !site.MyUser.IsValid() {
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
