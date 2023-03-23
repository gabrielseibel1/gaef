package auth

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MiddlewareGenerator struct {
	reader           tokenReader
	contextUserIDKey string
	contextTokenKey  string
}

type tokenReader interface {
	ReadToken(ctx context.Context, token string) (string, error)
}

func NewMiddlewareGenerator(reader tokenReader, contextUserIDKey, contextTokenKey string) MiddlewareGenerator {
	return MiddlewareGenerator{reader: reader, contextUserIDKey: contextUserIDKey, contextTokenKey: contextTokenKey}
}

func (g MiddlewareGenerator) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) <= len("Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorUnauthorized)
			return
		}
		token := authHeader[len("Bearer "):]

		userID, err := g.reader.ReadToken(ctx, token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorUnauthorized)
			return
		}

		ctx.Set(g.contextTokenKey, token)
		ctx.Set(g.contextUserIDKey, userID)

		ctx.Next()
	}
}

var errorUnauthorized = gin.H{"error": "unauthorized"}
