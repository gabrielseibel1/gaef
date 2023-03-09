package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gabrielseibel1/gaef-user-service/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// dependencies

type Creator interface {
	Create(user *domain.User, password string, ctx context.Context) (string, error)
}
type Loginer interface {
	Login(email, password string, ctx context.Context) (string, error)
}
type Reader interface {
	Read(id string, ctx context.Context) (*domain.User, error)
}
type Updater interface {
	Update(user *domain.User, ctx context.Context) (*domain.User, error)
}
type Deleter interface {
	Delete(id string, ctx context.Context) error
}

// implementation

type Handler struct {
	creator   Creator
	loginer   Loginer
	reader    Reader
	updater   Updater
	deleter   Deleter
	jwtSecret []byte
}

const jwtTTL = time.Hour * 24 * 7

func New(creator Creator, loginer Loginer, reader Reader, updater Updater, deleter Deleter, jwtSecret []byte) *Handler {
	return &Handler{
		creator:   creator,
		loginer:   loginer,
		reader:    reader,
		updater:   updater,
		deleter:   deleter,
		jwtSecret: jwtSecret,
	}
}

func (sh Handler) JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) <= len("Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}
		authHeader = authHeader[len("Bearer "):]

		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
			return sh.jwtSecret, nil
		})
		fmt.Println(token)
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		tokenUserID := claims["sub"].(string)

		paramUserID := ctx.Param("id")
		if paramUserID != "" && tokenUserID != paramUserID {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx.Set("AuthenticatedUserID", tokenUserID)

		ctx.Next()
	}
}

func (sh Handler) Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var json struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"` // TODO: password requirements validation
		}
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user := &domain.User{
			Email: json.Email,
			Name:  json.Name,
		}
		id, err := sh.creator.Create(user, json.Password, ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func (sh Handler) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var json struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized"})
			return
		}

		id, err := sh.loginer.Login(json.Email, json.Password, ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := sh.reader.Read(id, ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"name":  user.Name,
			"email": user.Email,
			"sub":   user.ID,
			"exp":   time.Now().Add(jwtTTL).Unix(),
		})
		tokenString, err := token.SignedString(sh.jwtSecret)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

func (sh Handler) GetIDFromToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString("AuthenticatedUserID")
		ctx.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func (sh Handler) GetUserFromID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		user, err := sh.reader.Read(id, ctx)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func (sh Handler) UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var providedUser domain.User
		if err := ctx.ShouldBindJSON(&providedUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if providedUser.ID != ctx.GetString("AuthenticatedUserID") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		updatedUser, err := sh.updater.Update(&providedUser, ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"user": updatedUser})
	}
}

func (sh Handler) DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString("AuthenticatedUserID")

		err := sh.deleter.Delete(id, ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deleted user %s", id)})
	}
}
