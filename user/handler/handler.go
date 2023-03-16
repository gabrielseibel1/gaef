package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gabrielseibel1/gaef/user/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// dependencies

type Creator interface {
	Create(user *domain.User, password string, ctx context.Context) (string, error)
}
type Loginer interface {
	Login(email, password string, ctx context.Context) (*domain.User, error)
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
const paramKeyAuthenticatedUserID = "AuthenticatedUserID"

var messageErrorUnauthorized = gin.H{"error": "unauthorized"}
var messageErrorMissingAuthorizationHeader = gin.H{"error": "missing authorization header"}
var messageErrorInvalidToken = gin.H{"error": "invalid or expired token"}
var messageErrorUserNotFound = gin.H{"error": "user not found"}
var messageErrorMissingUserData = gin.H{"error": "missing user data"}

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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, messageErrorMissingAuthorizationHeader)
			return
		}
		authHeader = authHeader[len("Bearer "):]

		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
			return sh.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, messageErrorInvalidToken)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, messageErrorInvalidToken)
			return
		}
		tokenUserID, ok := claims["sub"].(string)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, messageErrorInvalidToken)
			return
		}

		paramUserID := ctx.Param("id")
		if paramUserID != "" && tokenUserID != paramUserID {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		ctx.Set(paramKeyAuthenticatedUserID, tokenUserID)

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
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}
		user := &domain.User{
			Email: json.Email,
			Name:  json.Name,
		}
		id, err := sh.creator.Create(user, json.Password, ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is taken"})
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
			ctx.JSON(http.StatusBadRequest, messageErrorUnauthorized)
			return
		}

		user, err := sh.loginer.Login(json.Email, json.Password, ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

func (sh Handler) GetIDFromToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString(paramKeyAuthenticatedUserID)
		ctx.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func (sh Handler) GetUserFromID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString(paramKeyAuthenticatedUserID)
		user, err := sh.reader.Read(id, ctx)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func (sh Handler) UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString(paramKeyAuthenticatedUserID)

		var providedUser domain.User
		if err := ctx.ShouldBindJSON(&providedUser); err != nil {
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}

		if !(id == providedUser.ID) {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		updatedUser, err := sh.updater.Update(&providedUser, ctx)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"user": updatedUser})
	}
}

func (sh Handler) DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString(paramKeyAuthenticatedUserID)

		err := sh.deleter.Delete(id, ctx)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deleted user %s", id)})
	}
}
