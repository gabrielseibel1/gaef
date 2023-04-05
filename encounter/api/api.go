package api

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/encounter/server"
	"github.com/gabrielseibel1/gaef/types"
	"net/http"
)

type Result struct {
	Status int
	Name   string
	Value  any
}

func (r Result) S() int {
	return r.Status
}

func (r Result) K() string {
	return r.Name
}

func (r Result) V() any {
	return r.Value
}

type API struct {
	encounterCreator     EncounterCreator
	leaderChecker        LeaderChecker
	encounterReader      EncounterReader
	userEncountersReader UserEncountersReader
	encounterUpdater     EncounterUpdater
	encounterDeleter     EncounterDeleter
	encounterConfirmer   EncounterConfirmer
	encounterDecliner    EncounterDecliner
}

func New(leaderChecker LeaderChecker, encounterCreator EncounterCreator, encounterReader EncounterReader, userEncountersReader UserEncountersReader, encounterUpdater EncounterUpdater, encounterDeleter EncounterDeleter, encounterConfirmer EncounterConfirmer, encounterDecliner EncounterDecliner) API {
	return API{
		leaderChecker:        leaderChecker,
		encounterCreator:     encounterCreator,
		encounterReader:      encounterReader,
		userEncountersReader: userEncountersReader,
		encounterUpdater:     encounterUpdater,
		encounterDeleter:     encounterDeleter,
		encounterConfirmer:   encounterConfirmer,
		encounterDecliner:    encounterDecliner,
	}
}

func (a API) CreateEncounter(ctx context.Context, token string, e types.Encounter) server.Result {
	ok, err := a.userIsLeaderRemote(ctx, token, e)
	if err != nil {
		return errResult(http.StatusUnauthorized, err)
	}
	if !ok {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	id, err := a.encounterCreator.CreateEncounter(ctx, e)
	if err != nil {
		return errResult(http.StatusUnprocessableEntity, err)
	}
	return okResult(idName, id)
}

func (a API) ReadUserEncounters(ctx context.Context, userID string) server.Result {
	encs, err := a.userEncountersReader.ReadUserEncounters(ctx, userID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}
	return okResult(encountersName, encs)
}

func (a API) ReadEncounter(ctx context.Context, userID, encID string) server.Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}

	if !userIsInvited(enc, userID) {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	return okResult(encounterName, enc)
}

func (a API) UpdateEncounter(ctx context.Context, userID string, encID string, e types.Encounter) server.Result {
	// TODO test
	if encID != e.ID {
		return errResult(http.StatusUnprocessableEntity, errors.New("cannot edit id"))
	}

	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}

	if !userIsLeader(enc, userID) {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	enc, err = a.encounterUpdater.UpdateEncounter(ctx, e)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}
	return okResult(encounterName, enc)
}

func (a API) DeleteEncounter(ctx context.Context, userID, encID string) server.Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}

	if !userIsLeader(enc, userID) {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	if err := a.encounterDeleter.DeleteEncounter(ctx, encID); err != nil {
		return errResult(http.StatusNotFound, err)
	}
	return okResult(idName, encID)
}

func (a API) ConfirmEncounter(ctx context.Context, userID string, encID string) server.Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}

	user, err := getInvitedUser(enc, userID)
	if err != nil {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	if userIsConfirmed(enc, userID) {
		return errResult(http.StatusUnprocessableEntity, errors.New("user is already confirmed"))
	}

	if err := a.encounterConfirmer.ConfirmEncounter(ctx, encID, user); err != nil {
		return errResult(http.StatusNotFound, err)
	}
	return okResult(idName, encID)
}

func (a API) DeclineEncounter(ctx context.Context, userID string, encID string) server.Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(http.StatusNotFound, err)
	}

	_, err = getInvitedUser(enc, userID)
	if err != nil {
		return errResult(http.StatusUnauthorized, errUnauthorized)
	}

	if !userIsConfirmed(enc, userID) {
		return errResult(http.StatusUnprocessableEntity, errors.New("user is not confirmed"))
	}

	if err := a.encounterDecliner.DeclineEncounter(ctx, encID, userID); err != nil {
		return errResult(http.StatusNotFound, err)
	}
	return okResult(idName, encID)
}

type EncounterCreator interface {
	CreateEncounter(ctx context.Context, e types.Encounter) (string, error)
}

type LeaderChecker interface {
	IsGroupLeader(ctx context.Context, token string, groupID string) (bool, error)
}

type EncounterReader interface {
	ReadEncounter(ctx context.Context, id string) (types.Encounter, error)
}

type UserEncountersReader interface {
	ReadUserEncounters(ctx context.Context, userID string) ([]types.Encounter, error)
}

type EncounterUpdater interface {
	UpdateEncounter(ctx context.Context, e types.Encounter) (types.Encounter, error)
}

type EncounterDeleter interface {
	DeleteEncounter(ctx context.Context, id string) error
}

type EncounterConfirmer interface {
	ConfirmEncounter(ctx context.Context, encID string, user types.User) error
}

type EncounterDecliner interface {
	DeclineEncounter(ctx context.Context, encID, userID string) error
}

func (a API) userIsLeaderRemote(ctx context.Context, token string, enc types.Encounter) (bool, error) {
	var anyErr error
	for _, group := range enc.Groups {
		ok, err := a.leaderChecker.IsGroupLeader(ctx, token, group.ID)
		if err != nil {
			anyErr = err
			continue
		}
		if ok {
			return true, nil
		}
	}
	return false, anyErr
}

func getInvitedUser(enc types.Encounter, userID string) (types.User, error) {
	for _, invitedUser := range enc.InvitedUsers {
		if invitedUser.ID == userID {
			return invitedUser, nil
		}
	}
	return types.User{}, errors.New("user is not invited")
}

func userIsInvited(enc types.Encounter, userID string) bool {
	for _, invitedUser := range enc.InvitedUsers {
		if invitedUser.ID == userID {
			return true
		}
	}
	return false
}

func userIsConfirmed(enc types.Encounter, userID string) bool {
	for _, confirmedUser := range enc.ConfirmedUsers {
		if confirmedUser.ID == userID {
			return true
		}
	}
	return false
}

func userIsLeader(enc types.Encounter, userID string) bool {
	for _, group := range enc.Groups {
		for _, leader := range group.Leaders {
			if leader.ID == userID {
				return true
			}
		}
	}
	return false
}

func okResult(name string, value any) Result {
	return Result{Status: http.StatusOK, Name: name, Value: value}
}

func errResult(status int, err error) Result {
	return Result{Status: status, Name: errorName, Value: err.Error()}
}

var (
	errUnauthorized = errors.New("unauthorized")
)

var (
	idName         = "id"
	errorName      = "error"
	encounterName  = "encounter"
	encountersName = "encounters"
)
