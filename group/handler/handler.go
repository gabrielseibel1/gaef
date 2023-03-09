package handler

import (
	"context"
	"errors"
	"fmt"
	"gaef-group-service/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authenticator    Authenticator
	leaderChecker    LeaderChecker
	groupCreator     GroupCreator
	userGroupsReader UserGroupsReader
	groupReader      GroupReader
	groupUpdater     GroupUpdater
	groupDeleter     GroupDeleter
}

func New(
	authenticator Authenticator,
	leaderChecker LeaderChecker,
	groupCreator GroupCreator,
	userGroupsReader UserGroupsReader,
	groupReader GroupReader,
	groupUpdater GroupUpdater,
	groupDeleter GroupDeleter,
) Handler {
	return Handler{
		authenticator:    authenticator,
		leaderChecker:    leaderChecker,
		groupCreator:     groupCreator,
		userGroupsReader: userGroupsReader,
		groupReader:      groupReader,
		groupUpdater:     groupUpdater,
		groupDeleter:     groupDeleter,
	}
}

func (h Handler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) <= len("Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}
		token := authHeader[len("Bearer "):]

		userID, err := h.authenticator.Authenticate(ctx, token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorMessageUnauthorized)
			return
		}

		ctx.Set("userID", userID)

		ctx.Next()
	}
}

func (h Handler) OnlyLeadersMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		groupID := ctx.Param("id")
		userID := ctx.GetString("userID")

		ok, err := h.leaderChecker.IsLeader(ctx, userID, groupID)
		if err != nil || !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorMessageUnauthorized)
			return
		}

		ctx.Next()
	}
}

func (h Handler) CreateGroupHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var group domain.Group
		if err := ctx.ShouldBindJSON(&group); err != nil {
			ctx.JSON(http.StatusBadRequest, ginErrorMessage(err))
			return
		}

		group, err := h.groupCreator.CreateGroup(ctx, group)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ginErrorMessage(err))
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"group": group})
	}
}

func (h Handler) ReadAllGroupsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")

		groups, err := h.userGroupsReader.ReadGroups(ctx, userID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ginErrorMessage(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"groups": groups})
	}
}

func (h Handler) ReadGroupHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		groupID := ctx.Param("id")

		group, err := h.groupReader.ReadGroup(ctx, groupID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ginErrorMessage(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"group": group})
	}
}

func (h Handler) UpdateGroupHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		groupID := ctx.Param("id")

		var group domain.Group
		if err := ctx.ShouldBindJSON(&group); err != nil {
			ctx.JSON(http.StatusBadRequest, ginErrorMessage(err))
			return
		}
		if groupID != group.ID {
			ctx.JSON(http.StatusBadRequest, ginErrorMessage(errors.New("group id cannot be updated")))
			return
		}

		group, err := h.groupUpdater.UpdateGroup(ctx, group)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ginErrorMessage(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"group": group})
	}
}

func (h Handler) DeleteGroupHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		groupID := ctx.Param("id")

		err := h.groupDeleter.DeleteGroup(ctx, groupID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ginErrorMessage(err))
		}

		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deleted group %s", groupID)})
	}
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (string, error)
}
type LeaderChecker interface {
	IsLeader(ctx context.Context, userID string, groupID string) (bool, error)
}
type GroupCreator interface {
	CreateGroup(ctx context.Context, group domain.Group) (domain.Group, error)
}
type UserGroupsReader interface {
	ReadGroups(ctx context.Context, userID string) ([]domain.Group, error)
}
type GroupReader interface {
	ReadGroup(ctx context.Context, id string) (domain.Group, error)
}
type GroupUpdater interface {
	UpdateGroup(ctx context.Context, group domain.Group) (domain.Group, error)
}
type GroupDeleter interface {
	DeleteGroup(ctx context.Context, id string) error
}

var errorMessageUnauthorized gin.H = gin.H{"error": "unauthorized"}

func ginErrorMessage(err error) gin.H {
	return gin.H{"error": err.Error()}
}
