package handler

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// dependencies

type PasswordHasher interface {
	GenerateFromPassword(password string) (string, error)
}
type PasswordVerifier interface {
	CompareHashAndPassword(hashedPassword, password string) error
}
type Creator interface {
	Create(ctx context.Context, user types.User) (string, error)
}
type ByEmailReader interface {
	ReadSensitiveByEmail(ctx context.Context, email string) (types.User, error)
}
type ByIDReader interface {
	ReadByID(ctx context.Context, id string) (types.User, error)
}
type Updater interface {
	Update(ctx context.Context, user types.User) error
}
type Deleter interface {
	Delete(ctx context.Context, id string) error
}

// implementation

type Handler struct {
	hasher        PasswordHasher
	verifier      PasswordVerifier
	creator       Creator
	byIDReader    ByIDReader
	byEmailReader ByEmailReader
	updater       Updater
	deleter       Deleter
	jwtSecret     []byte
}

const jwtTTL = time.Hour * 24 * 7
const paramKeyAuthenticatedUserID = "AuthenticatedUserID"

var messageErrorUnauthorized = gin.H{"error": "unauthorized"}
var messageErrorMissingAuthorizationHeader = gin.H{"error": "missing authorization header"}
var messageErrorInvalidToken = gin.H{"error": "invalid or expired token"}
var messageErrorUserNotFound = gin.H{"error": "user not found"}
var messageErrorMissingUserData = gin.H{"error": "missing user data"}

func New(
	hasher PasswordHasher,
	verifier PasswordVerifier,
	creator Creator,
	byIDReader ByIDReader,
	byEmailReader ByEmailReader,
	updater Updater,
	deleter Deleter,
	jwtSecret []byte,
) *Handler {
	return &Handler{
		hasher:        hasher,
		verifier:      verifier,
		creator:       creator,
		byIDReader:    byIDReader,
		byEmailReader: byEmailReader,
		updater:       updater,
		deleter:       deleter,
		jwtSecret:     jwtSecret,
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
		user := types.User{
			Name:     json.Name,
			Email:    json.Email,
			Password: json.Password,
		}

		_, err := sh.byEmailReader.ReadSensitiveByEmail(ctx, user.Email)
		if err == nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "email is taken"})
			return
		}

		user.HashedPassword, err = sh.hasher.GenerateFromPassword(user.Password)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "bad password"})
			return
		}

		id, err := sh.creator.Create(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "email is taken"})
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
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}

		u, err := sh.byEmailReader.ReadSensitiveByEmail(ctx, json.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		err = sh.verifier.CompareHashAndPassword(u.HashedPassword, json.Password)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"name":  u.Name,
			"email": u.Email,
			"sub":   u.ID,
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
		user, err := sh.byIDReader.ReadByID(ctx, id)
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

		var json struct {
			ID   string `json:"id" binding:"required"`
			Name string `json:"name" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}
		user := types.User{ID: json.ID, Name: json.Name}

		if !(id == user.ID) {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		err := sh.updater.Update(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func (sh Handler) DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetString(paramKeyAuthenticatedUserID)

		err := sh.deleter.Delete(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deleted user %s", id)})
	}
}
