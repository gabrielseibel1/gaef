package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var (
	Page = "page"
	EPID = "epid"
)

type API struct {
	epCreator           encounterProposalCreator
	pagedEPsReader      pagedEncounterProposalsReader
	byGroupIDsEPsReader byGroupIDsEPsReader
	byIDEPReader        byIDEncounterProposalReader
	epUpdater           encounterProposalUpdater
	epDeleter           encounterProposalDeleter
	appAppender         applicationAppender
	leadingGroupsLister leadingGroupsLister
	leaderChecker       groupLeaderChecker
}

type encounterProposalCreator interface {
	Create(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error)
}
type pagedEncounterProposalsReader interface {
	ReadPaged(ctx context.Context, page int) ([]types.EncounterProposal, error)
}
type byGroupIDsEPsReader interface {
	ReadByGroupIDs(ctx context.Context, groupIDs []string) ([]types.EncounterProposal, error)
}
type byIDEncounterProposalReader interface {
	ReadByID(ctx context.Context, id string) (types.EncounterProposal, error)
}
type encounterProposalUpdater interface {
	Update(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error)
}
type encounterProposalDeleter interface {
	Delete(ctx context.Context, id string) error
}
type applicationAppender interface {
	Append(ctx context.Context, epID string, app types.Application) error
}
type leadingGroupsLister interface {
	LeadingGroups(ctx context.Context, token string) ([]types.Group, error)
}
type groupLeaderChecker interface {
	IsGroupLeader(ctx context.Context, token string, groupID string) (bool, error)
}

func New(
	epCreator encounterProposalCreator,
	pagedEPsReader pagedEncounterProposalsReader,
	byGroupIDsEPsReader byGroupIDsEPsReader,
	byIDEPReader byIDEncounterProposalReader,
	epUpdater encounterProposalUpdater,
	epDeleter encounterProposalDeleter,
	appAppender applicationAppender,
	leadingGroupsLister leadingGroupsLister,
	leaderChecker groupLeaderChecker,
) API {
	return API{
		epCreator:           epCreator,
		pagedEPsReader:      pagedEPsReader,
		byGroupIDsEPsReader: byGroupIDsEPsReader,
		byIDEPReader:        byIDEPReader,
		epUpdater:           epUpdater,
		epDeleter:           epDeleter,
		appAppender:         appAppender,
		leadingGroupsLister: leadingGroupsLister,
		leaderChecker:       leaderChecker,
	}
}

func (api API) EPCreatorGroupLeaderCheckerMiddleware() gin.HandlerFunc {
	return jsonMiddleware(func(ctx *gin.Context) status {

		epID := ctx.Param(EPID)
		ep, err := api.byIDEPReader.ReadByID(ctx, epID)
		if err != nil {
			return apiErrorUnauthorized
		}

		token := ctx.GetString(authenticatedUserToken)
		isLeader, err := api.leaderChecker.IsGroupLeader(ctx, token, ep.Creator.ID)
		if err != nil || !isLeader {
			return apiErrorUnauthorized
		}

		return status{}

	})
}

func (api API) EPCreationHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		var ep types.EncounterProposal
		if err := ctx.ShouldBindJSON(&ep); err != nil {
			return er(http.StatusBadRequest, err)
		}

		ep, err := api.epCreator.Create(ctx, ep)
		if err != nil {
			return er(http.StatusConflict, err)
		}

		return result{
			s: status{
				code: http.StatusCreated,
			},
			r: resource{
				k: encounterProposal,
				v: ep,
			},
		}

	})
}

func (api API) EPReadingAllHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		page, err := strconv.Atoi(ctx.Param(Page))
		if err != nil {
			return er(http.StatusBadRequest, err)
		}

		eps, err := api.pagedEPsReader.ReadPaged(ctx, page)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(encounterProposalSlice, eps)

	})
}

func (api API) EPReadingByUserHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		token := ctx.GetString(authenticatedUserToken)
		leadingGroups, err := api.leadingGroupsLister.LeadingGroups(ctx, token)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		ids := make([]string, len(leadingGroups))
		for i, g := range leadingGroups {
			ids[i] = g.ID
		}

		eps, err := api.byGroupIDsEPsReader.ReadByGroupIDs(ctx, ids)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(encounterProposalSlice, eps)

	})
}

func (api API) EPReadingByIDHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		id := ctx.Param(EPID)
		ep, err := api.byIDEPReader.ReadByID(ctx, id)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(encounterProposal, ep)

	})
}

func (api API) EPUpdateHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		var ep types.EncounterProposal
		if err := ctx.ShouldBindJSON(&ep); err != nil {
			return er(http.StatusBadRequest, err)
		}
		if ep.ID != ctx.Param(EPID) {
			return er(http.StatusUnprocessableEntity, errors.New("cannot update id"))
		}

		ep, err := api.epUpdater.Update(ctx, ep)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(encounterProposal, ep)

	})
}

func (api API) EPDeletionHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		id := ctx.Param(EPID)
		err := api.epDeleter.Delete(ctx, id)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(message, fmt.Sprintf("deleted encounter proposal %s", id))

	})
}

func (api API) AppCreationHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		var app types.Application
		if err := ctx.ShouldBindJSON(&app); err != nil {
			return er(http.StatusBadRequest, err)
		}

		// check if user is a leader of the applicant group
		token := ctx.GetString(authenticatedUserToken)
		isLeader, err := api.leaderChecker.IsGroupLeader(ctx, token, app.Creator.ID)
		if err != nil {
			return er(http.StatusUnauthorized, err)
		}
		if !isLeader {
			return er(http.StatusUnauthorized, errors.New("user is not a leader of applicant group"))
		}

		// update EP with the new application
		epID := ctx.Param(EPID)
		err = api.appAppender.Append(ctx, epID, app)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(message, fmt.Sprintf("applied for %s", epID))

	})
}

type result struct {
	s status
	r resource
}

type status struct {
	code int
	err  error
}

type resource struct {
	k string
	v any
}

func jsonMiddleware(f func(*gin.Context) status) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := f(ctx)
		if status.err != nil {
			ctx.AbortWithStatusJSON(status.code, gin.H{"error": status.err.Error()})
			return
		}
		ctx.Next()
	}
}

func jsonHandler(f func(*gin.Context) result) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := f(ctx)
		if r.s.err != nil {
			ctx.JSON(r.s.code, gin.H{"error": r.s.err.Error()})
			return
		}
		ctx.JSON(r.s.code, gin.H{r.r.k: r.r.v})
	}
}

func ok(resourceName string, resourceValue any) result {
	return result{
		s: status{
			code: http.StatusOK,
		},
		r: resource{
			k: resourceName,
			v: resourceValue,
		},
	}
}

func er(code int, err error) result {
	return result{
		s: status{
			code: code,
			err:  err,
		},
	}
}

var (
	authenticatedUserToken = "token"
	encounterProposal      = "encounterProposal"
	encounterProposalSlice = "encounterProposals"
	message                = "message"
)

var (
	apiErrorUnauthorized = status{
		code: http.StatusUnauthorized,
		err:  errors.New("unauthorized"),
	}
)
