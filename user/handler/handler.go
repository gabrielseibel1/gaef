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
	Create(ctx context.Context, user types.UserWithHashedPassword) (string, error)
}
type ByEmailReader interface {
	ReadSensitiveByEmail(ctx context.Context, email string) (types.UserWithHashedPassword, error)
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
type UpdateMessenger interface {
	SendUserUpdatedMessage(ctx context.Context, user types.User) error
}
type DeleteMessenger interface {
	SendUserDeletedMessage(ctx context.Context, userID string) error
}

// implementation

type Handler struct {
	hasher              PasswordHasher
	verifier            PasswordVerifier
	creator             Creator
	byIDReader          ByIDReader
	byEmailReader       ByEmailReader
	updater             Updater
	deleter             Deleter
	userUpdateMessenger UpdateMessenger
	userDeleteMessenger DeleteMessenger
	jwtSecret           []byte
}

func New(
	hasher PasswordHasher,
	verifier PasswordVerifier,
	creator Creator,
	byIDReader ByIDReader,
	byEmailReader ByEmailReader,
	updater Updater,
	deleter Deleter,
	userUpdateMessenger UpdateMessenger,
	userDeleteMessenger DeleteMessenger,
	jwtSecret []byte,
) *Handler {
	return &Handler{
		hasher:              hasher,
		verifier:            verifier,
		creator:             creator,
		byIDReader:          byIDReader,
		byEmailReader:       byEmailReader,
		updater:             updater,
		deleter:             deleter,
		userUpdateMessenger: userUpdateMessenger,
		userDeleteMessenger: userDeleteMessenger,
		jwtSecret:           jwtSecret,
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
			User     types.User `json:"user" binding:"required"`
			Password string     `json:"password" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}
		user := types.UserWithHashedPassword{User: json.User}

		_, err := sh.byEmailReader.ReadSensitiveByEmail(ctx, user.Email)
		if err == nil {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "email is taken"})
			return
		}

		user.HashedPassword, err = sh.hasher.GenerateFromPassword(json.Password)
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

		var user types.User
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, messageErrorMissingUserData)
			return
		}

		if !(id == user.ID) {
			ctx.JSON(http.StatusUnauthorized, messageErrorUnauthorized)
			return
		}

		err := sh.updater.Update(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusNotFound, messageErrorUserNotFound)
			return
		}

		err = sh.userUpdateMessenger.SendUserUpdatedMessage(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, messageErrorMessageBroker)
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

		err = sh.userDeleteMessenger.SendUserDeletedMessage(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, messageErrorMessageBroker)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deleted user %s", id)})
	}
}

const jwtTTL = time.Hour * 24 * 7
const paramKeyAuthenticatedUserID = "AuthenticatedUserID"

var (
	messageErrorUnauthorized               = gin.H{"error": "unauthorized"}
	messageErrorMissingAuthorizationHeader = gin.H{"error": "missing authorization header"}
	messageErrorInvalidToken               = gin.H{"error": "invalid or expired token"}
	messageErrorUserNotFound               = gin.H{"error": "user not found"}
	messageErrorMissingUserData            = gin.H{"error": "missing user data"}
	messageErrorMessageBroker              = gin.H{"error": "unable to send message to broker"}
)
