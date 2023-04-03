package api

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/types"
)

type Result struct {
	Name  string
	Value any
	Err   error
}

type API struct {
	encounterCreator     EncounterCreator
	leaderChecker        LeaderChecker
	encounterReader      EncounterReader
	userEncountersReader UserEncountersReader
	encounterUpdater     EncounterUpdater
	encounterDeleter     EncounterDeleter
	encounterConfirmer   EncounterConfirmer
}

func New(
	encounterCreator EncounterCreator,
	leaderChecker LeaderChecker,
	encounterReader EncounterReader,
	userEncountersReader UserEncountersReader,
	encounterUpdater EncounterUpdater,
	encounterDeleter EncounterDeleter,
	encounterConfirmer EncounterConfirmer,
) API {
	return API{
		encounterCreator:     encounterCreator,
		leaderChecker:        leaderChecker,
		encounterReader:      encounterReader,
		userEncountersReader: userEncountersReader,
		encounterUpdater:     encounterUpdater,
		encounterDeleter:     encounterDeleter,
		encounterConfirmer:   encounterConfirmer,
	}
}

func (a API) CreateEncounter(ctx context.Context, token string, e types.Encounter) Result {
	ok, err := a.userIsLeaderRemote(ctx, token, e)
	if err != nil {
		return errResult(err)
	}
	if !ok {
		return errResult(errUnauthorized)
	}

	id, err := a.encounterCreator.CreateEncounter(ctx, e)
	if err != nil {
		return errResult(err)
	}
	return okResult(idName, id)
}

func (a API) ReadUserEncounters(ctx context.Context, userID string) Result {
	encs, err := a.userEncountersReader.ReadUserEncounters(ctx, userID)
	if err != nil {
		return errResult(err)
	}
	return okResult(encountersName, encs)
}

func (a API) ReadEncounter(ctx context.Context, encID, userID string) Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(err)
	}

	if !userIsInvited(enc, userID) {
		return errResult(errUnauthorized)
	}

	return okResult(encounterName, enc)
}

func (a API) UpdateEncounter(ctx context.Context, userID string, e types.Encounter) Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, e.ID)
	if err != nil {
		return errResult(err)
	}

	if !userIsLeader(enc, userID) {
		return errResult(errUnauthorized)
	}

	enc, err = a.encounterUpdater.UpdateEncounter(ctx, e)
	if err != nil {
		return errResult(err)
	}
	return okResult(encounterName, enc)
}

func (a API) DeleteEncounter(ctx context.Context, userID, encID string) Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(err)
	}

	if !userIsLeader(enc, userID) {
		return errResult(errUnauthorized)
	}

	if err := a.encounterDeleter.DeleteEncounter(ctx, encID); err != nil {
		return errResult(err)
	}
	return okResult(idName, encID)
}

func (a API) ConfirmEncounter(ctx context.Context, encID string, userID string) Result {
	enc, err := a.encounterReader.ReadEncounter(ctx, encID)
	if err != nil {
		return errResult(err)
	}

	user, err := getInvitedUser(enc, userID)
	if err != nil {
		return errResult(errUnauthorized)
	}

	if userIsConfirmed(enc, userID) {
		return errResult(errors.New("user is already confirmed"))
	}

	if err := a.encounterConfirmer.ConfirmEncounter(ctx, encID, user); err != nil {
		return errResult(err)
	}
	return okResult(idName, encID)
}

type EncounterCreator interface {
	CreateEncounter(ctx context.Context, e types.Encounter) (string, error)
}

type LeaderChecker interface {
	IsLeader(ctx context.Context, token string, groupID string) (bool, error)
}

type EncounterReader interface {
	ReadEncounter(ctx context.Context, id string) (types.Encounter, error)
}

type UserEncountersReader interface {
	ReadUserEncounters(ctx context.Context, token string) ([]types.Encounter, error)
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

func (a API) userIsLeaderRemote(ctx context.Context, token string, enc types.Encounter) (bool, error) {
	var anyErr error
	for _, group := range enc.Groups {
		ok, err := a.leaderChecker.IsLeader(ctx, token, group.ID)
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
	return Result{Name: name, Value: value}
}

func errResult(err error) Result {
	return Result{Name: errorName, Value: err.Error(), Err: err}
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
