package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gabrielseibel1/gaef/group/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authenticator             Authenticator
	leaderChecker             LeaderChecker
	groupCreator              GroupCreator
	participatingGroupsReader ParticipatingGroupsReader
	leadingGroupsReader       LeadingGroupsReader
	groupReader               GroupReader
	groupUpdater              GroupUpdater
	groupDeleter              GroupDeleter
}

func New(
	leaderChecker LeaderChecker,
	groupCreator GroupCreator,
	participatingGroupsReader ParticipatingGroupsReader,
	leadingGroupsReader LeadingGroupsReader,
	groupReader GroupReader,
	groupUpdater GroupUpdater,
	groupDeleter GroupDeleter,
) Handler {
	return Handler{
		leaderChecker:             leaderChecker,
		groupCreator:              groupCreator,
		participatingGroupsReader: participatingGroupsReader,
		leadingGroupsReader:       leadingGroupsReader,
		groupReader:               groupReader,
		groupUpdater:              groupUpdater,
		groupDeleter:              groupDeleter,
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

func (h Handler) ReadParticipatingGroupsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")

		groups, err := h.participatingGroupsReader.ReadParticipatingGroups(ctx, userID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ginErrorMessage(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"groups": groups})
	}
}

func (h Handler) ReadLeadingGroupsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetString("userID")

		groups, err := h.leadingGroupsReader.ReadLeadingGroups(ctx, userID)
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
type ParticipatingGroupsReader interface {
	ReadParticipatingGroups(ctx context.Context, userID string) ([]domain.Group, error)
}
type LeadingGroupsReader interface {
	ReadLeadingGroups(ctx context.Context, userID string) ([]domain.Group, error)
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

var errorMessageUnauthorized = gin.H{"error": "unauthorized"}

func ginErrorMessage(err error) gin.H {
	return gin.H{"error": err.Error()}
}
