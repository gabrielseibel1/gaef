package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Generator struct {
	authenticator    authenticator
	contextUserIDKey string
}

type authenticator interface {
	GetAuthenticatedUserID(ctx context.Context, token string) (string, error)
}

func New(authenticator authenticator, contextUserIDKey string) Generator {
	return Generator{authenticator: authenticator, contextUserIDKey: contextUserIDKey}
}

func (g Generator) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) <= len("Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorUnauthorized)
			return
		}
		token := authHeader[len("Bearer "):]

		userID, err := g.authenticator.GetAuthenticatedUserID(ctx, token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorUnauthorized)
			return
		}

		ctx.Set(g.contextUserIDKey, userID)

		ctx.Next()
	}
}

var errorUnauthorized = gin.H{"error": "unauthorized"}
