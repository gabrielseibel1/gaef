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
	Page                   = "page"
	EPID                   = "epid"
	AppID                  = "appid"
	AuthenticatedUserID    = "userID"
	AuthenticatedUserToken = "token"
)

type API struct {
	epCreator           encounterProposalCreator
	pagedEPsReader      pagedEncounterProposalsReader
	byGroupIDsEPsReader byGroupIDsEPsReader
	byIDEPReader        byIDEncounterProposalReader
	epUpdater           encounterProposalUpdater
	epDeleter           encounterProposalDeleter
	appAppender         applicationAppender
	appDeleter          applicationDeleter
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
	AppendApplication(ctx context.Context, epID string, app types.Application) error
}
type applicationDeleter interface {
	DeleteApplication(ctx context.Context, epID string, appID string) error
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
	appDeleter applicationDeleter,
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
		appDeleter:          appDeleter,
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

		userID := ctx.GetString(AuthenticatedUserID)
		isLeader := false
		for _, leader := range ep.Creator.Leaders {
			if leader.ID == userID {
				isLeader = true
			}
		}
		if !isLeader {
			return apiErrorUnauthorized
		}

		return next()

	})
}

func (api API) EPCreationHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		var ep types.EncounterProposal
		if err := ctx.ShouldBindJSON(&ep); err != nil {
			return er(http.StatusBadRequest, err)
		}

		// check that user sending request is leader of creator group
		token := ctx.GetString(AuthenticatedUserToken)
		isLeader, err := api.leaderChecker.IsGroupLeader(ctx, token, ep.Creator.ID)
		if err != nil || !isLeader {
			return result{s: apiErrorUnauthorized}
		}

		ep, err = api.epCreator.Create(ctx, ep)
		if err != nil {
			return er(http.StatusConflict, err)
		}

		return result{
			s: status{code: http.StatusCreated},
			r: resource{k: "id", v: ep.ID},
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

		token := ctx.GetString(AuthenticatedUserToken)
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

		// reset user input on applications field
		readEP, err := api.byIDEPReader.ReadByID(ctx, ep.ID)
		if err != nil {
			return er(http.StatusNotFound, err)
		}
		ep.Applications = readEP.Applications

		ep, err = api.epUpdater.Update(ctx, ep)
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
		token := ctx.GetString(AuthenticatedUserToken)
		isLeader, err := api.leaderChecker.IsGroupLeader(ctx, token, app.Applicant.ID)
		if err != nil {
			return er(http.StatusUnauthorized, err)
		}
		if !isLeader {
			return er(http.StatusUnauthorized, errors.New("user is not a leader of applicant group"))
		}

		// update EP with the new application
		epID := ctx.Param(EPID)
		err = api.appAppender.AppendApplication(ctx, epID, app)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(message, fmt.Sprintf("applied for encounter proposal %s", epID))

	})
}

func (api API) AppDeletionHandler() gin.HandlerFunc {
	return jsonHandler(func(ctx *gin.Context) result {

		epID := ctx.Param(EPID)
		appID := ctx.Param(AppID)

		err := api.appDeleter.DeleteApplication(ctx, epID, appID)
		if err != nil {
			return er(http.StatusNotFound, err)
		}

		return ok(message, fmt.Sprintf("deleted application %s of encounter proposal %s", appID, epID))

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

func next() status {
	return status{}
}

var (
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
