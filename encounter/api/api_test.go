package api_test

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/encounter/api"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func dummyAPIError(status int) api.Result {
	return api.Result{Status: status, Name: "error", Value: dummyError.Error()}
}

var (
	dummyError           = errors.New("dummy error")
	unauthorizedAPIError = api.Result{Status: http.StatusUnauthorized, Name: "error", Value: "unauthorized"}
	dummyCtx             = context.TODO()
	dummyToken           = "dummy-token"
	dummyID              = "dummy-id"
	dummyUser1           = types.User{
		ID:    "dummy-user-id-1",
		Name:  "dummy-user-name-1",
		Email: "dummy-user-email-1",
	}
	dummyUser2 = types.User{
		ID:    "dummy-user-id-2",
		Name:  "dummy-user-name-2",
		Email: "dummy-user-email-2",
	}
	dummyGroup1 = types.Group{
		ID:          "dummy-group-id-1",
		Name:        "dummy-group-name-1",
		PictureURL:  "dummy-group-picture-url-1",
		Description: "dummy-group-description-1",
		Members:     []types.User{dummyUser1},
		Leaders:     []types.User{dummyUser1},
	}
	dummyGroup2 = types.Group{
		ID:          "dummy-group-id-2",
		Name:        "dummy-group-name-2",
		PictureURL:  "dummy-group-picture-url-2",
		Description: "dummy-group-description-2",
		Members:     []types.User{dummyUser2},
		Leaders:     []types.User{dummyUser2},
	}
	dummyEncounter1 = types.Encounter{
		ID: "dummy-encounter-id-1",
		EncounterSpecification: types.EncounterSpecification{
			Name:        "dummy-encounter-name-1",
			Description: "dummy-encounter-description-1",
			Location: types.Location{
				Name:      "dummy-location-name-1",
				Latitude:  42.42,
				Longitude: 87.87,
			},
			Time: time.Now(),
		},
		Groups:         []types.Group{dummyGroup1, dummyGroup2},
		InvitedUsers:   []types.User{dummyUser1, dummyUser2},
		ConfirmedUsers: []types.User{dummyUser2},
	}
	dummyEncounter2 = types.Encounter{
		ID: "dummy-encounter-id-2",
		EncounterSpecification: types.EncounterSpecification{
			Name:        "dummy-encounter-name-2",
			Description: "dummy-encounter-description-2",
			Location: types.Location{
				Name:      "dummy-location-name-2",
				Latitude:  42.42,
				Longitude: 87.87,
			},
			Time: time.Now(),
		},
		Groups:         []types.Group{dummyGroup1, dummyGroup2},
		InvitedUsers:   []types.User{dummyUser1, dummyUser2},
		ConfirmedUsers: []types.User{dummyUser1, dummyUser2},
	}
	dummyEncounters = []types.Encounter{dummyEncounter1, dummyEncounter2}
)

type mocks struct {
	encounterCreator     api.EncounterCreator
	leaderChecker        api.LeaderChecker
	encounterReader      api.EncounterReader
	userEncountersReader api.UserEncountersReader
	encounterUpdater     api.EncounterUpdater
	encounterDeleter     api.EncounterDeleter
	encounterConfirmer   api.EncounterConfirmer
	encounterDecliner    api.EncounterDecliner
}

func apiFromMocks(m mocks) api.API {
	return api.New(m.leaderChecker, m.encounterCreator, m.encounterReader, m.userEncountersReader, m.encounterUpdater, m.encounterDeleter, m.encounterConfirmer, m.encounterDecliner)
}

func TestResult_S(t *testing.T) {
	r := api.Result{Status: 42}
	assert.Equal(t, 42, r.S())
}

func TestResult_K(t *testing.T) {
	r := api.Result{Name: "name"}
	assert.Equal(t, "name", r.K())
}

func TestResult_V(t *testing.T) {
	r := api.Result{Value: "val"}
	assert.Equal(t, "val", r.V())
}

func TestAPI_CreateEncounter(t *testing.T) {
	type args struct {
		ctx   context.Context
		token string
		e     types.Encounter
	}
	dummyArgs := args{
		ctx:   dummyCtx,
		token: dummyToken,
		e:     dummyEncounter1,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "create encounter ok",
			mocks: mocks{
				encounterCreator: &mockEncounterCreator{id: dummyID},
				leaderChecker:    &mockLeaderChecker{isLeader: true},
			},
			args: dummyArgs,
			want: api.Result{
				Status: http.StatusOK,
				Name:   "id",
				Value:  dummyID,
			},
			wantMocks: mocks{
				encounterCreator: &mockEncounterCreator{
					ctx: dummyCtx,
					enc: dummyEncounter1,
					id:  dummyID,
				},
				leaderChecker: &mockLeaderChecker{
					ctx:      dummyCtx,
					token:    dummyToken,
					groupID:  dummyEncounter1.Groups[0].ID,
					isLeader: true,
				},
			},
		},
		{
			name: "create encounter leader error",
			mocks: mocks{
				encounterCreator: &mockEncounterCreator{id: dummyID},
				leaderChecker:    &mockLeaderChecker{isLeader: true, err: dummyError},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusUnauthorized),
			wantMocks: mocks{
				encounterCreator: &mockEncounterCreator{
					id: dummyID,
				},
				leaderChecker: &mockLeaderChecker{
					ctx:      dummyCtx,
					token:    dummyToken,
					isLeader: true,
					err:      dummyError,
					groupID:  dummyEncounter1.Groups[1].ID,
				},
			},
		},
		{
			name: "create encounter leader false",
			mocks: mocks{
				encounterCreator: &mockEncounterCreator{id: dummyID},
				leaderChecker:    &mockLeaderChecker{isLeader: false},
			},
			args: dummyArgs,
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterCreator: &mockEncounterCreator{
					id: dummyID,
				},
				leaderChecker: &mockLeaderChecker{
					ctx:      dummyCtx,
					token:    dummyToken,
					isLeader: false,
					groupID:  dummyEncounter1.Groups[1].ID,
				},
			},
		},
		{
			name: "create encounter creator error",
			mocks: mocks{
				encounterCreator: &mockEncounterCreator{id: dummyID, err: dummyError},
				leaderChecker:    &mockLeaderChecker{isLeader: true},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusUnprocessableEntity),
			wantMocks: mocks{
				encounterCreator: &mockEncounterCreator{
					ctx: dummyCtx,
					enc: dummyEncounter1,
					id:  dummyID,
					err: dummyError,
				},
				leaderChecker: &mockLeaderChecker{
					ctx:      dummyCtx,
					token:    dummyToken,
					isLeader: true,
					groupID:  dummyEncounter1.Groups[0].ID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.CreateEncounter(tt.args.ctx, tt.args.token, tt.args.e),
				"CreateEncounter(%v, %v, %v)",
				tt.args.ctx,
				tt.args.token,
				tt.args.e,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_ReadUserEncounters(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		userID: dummyID,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name:  "read user encounters ok",
			mocks: mocks{userEncountersReader: &mockUserEncountersReader{encs: dummyEncounters}},
			args:  dummyArgs,
			want: api.Result{
				Status: http.StatusOK,
				Name:   "encounters",
				Value:  dummyEncounters,
			},
			wantMocks: mocks{
				userEncountersReader: &mockUserEncountersReader{
					ctx:    dummyCtx,
					userID: dummyID,
					encs:   dummyEncounters,
				},
			},
		},
		{
			name:  "read user encounters error",
			mocks: mocks{userEncountersReader: &mockUserEncountersReader{encs: dummyEncounters, err: dummyError}},
			args:  dummyArgs,
			want:  dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				userEncountersReader: &mockUserEncountersReader{
					ctx:    dummyCtx,
					userID: dummyID,
					encs:   dummyEncounters,
					err:    dummyError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.ReadUserEncounters(tt.args.ctx, tt.args.userID),
				"ReadUserEncounters(%v, %v)",
				tt.args.ctx,
				tt.args.userID,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_ReadEncounter(t *testing.T) {
	type args struct {
		ctx    context.Context
		encID  string
		userID string
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		encID:  dummyEncounter1.ID,
		userID: dummyUser1.ID,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "read encounter ok",
			mocks: mocks{
				encounterReader: &mockEncounterReader{enc: dummyEncounter1},
			},
			args: dummyArgs,
			want: api.Result{Status: http.StatusOK, Name: "encounter", Value: dummyEncounter1},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
			},
		},
		{
			name: "read encounter reader error",
			mocks: mocks{
				encounterReader: &mockEncounterReader{enc: dummyEncounter1, err: dummyError},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
					err: dummyError,
				},
			},
		},
		{
			name: "read encounter user not invited",
			mocks: mocks{
				encounterReader: &mockEncounterReader{enc: dummyEncounter1},
			},
			args: args{
				ctx:    dummyArgs.ctx,
				encID:  dummyArgs.encID,
				userID: dummyID,
			},
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.ReadEncounter(tt.args.ctx, tt.args.userID, tt.args.encID),
				"ReadEncounter(%v, %v, %v)",
				tt.args.ctx,
				tt.args.encID,
				tt.args.userID,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_UpdateEncounter(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		encID  string
		enc    types.Encounter
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		userID: dummyUser1.ID,
		encID:  dummyEncounter1.ID,
		enc:    dummyEncounter1,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "update encounter ok",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter2},
				encounterUpdater: &mockEncounterUpdater{retEnc: dummyEncounter2},
			},
			args: dummyArgs,
			want: api.Result{
				Status: http.StatusOK,
				Name:   "encounter",
				Value:  dummyEncounter2,
			},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter2,
				},
				encounterUpdater: &mockEncounterUpdater{
					ctx:    dummyCtx,
					rcvEnc: dummyEncounter1,
					retEnc: dummyEncounter2,
				},
			},
		},
		{
			name: "update encounter leader false",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter2},
				encounterUpdater: &mockEncounterUpdater{retEnc: dummyEncounter2},
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyID,
				encID:  dummyEncounter1.ID,
				enc:    dummyEncounter1,
			},
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter2,
				},
				encounterUpdater: &mockEncounterUpdater{
					retEnc: dummyEncounter2,
				},
			},
		},
		{
			name: "update encounter updater error",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter2},
				encounterUpdater: &mockEncounterUpdater{retEnc: dummyEncounter2, err: dummyError},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter2,
				},
				encounterUpdater: &mockEncounterUpdater{
					ctx:    dummyCtx,
					rcvEnc: dummyEncounter1,
					retEnc: dummyEncounter2,
					err:    dummyError,
				},
			},
		},
		{
			name: "update encounter reader error",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter2, err: dummyError},
				encounterUpdater: &mockEncounterUpdater{retEnc: dummyEncounter2},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter2,
					err: dummyError,
				},
				encounterUpdater: &mockEncounterUpdater{
					retEnc: dummyEncounter2,
				},
			},
		},
		{
			name: "update encounter edit id",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter2},
				encounterUpdater: &mockEncounterUpdater{retEnc: dummyEncounter2},
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyUser1.ID,
				encID:  dummyEncounter2.ID,
				enc:    dummyEncounter1,
			},
			want: api.Result{
				Status: http.StatusUnprocessableEntity,
				Name:   "error",
				Value:  "cannot edit id",
			},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					enc: dummyEncounter2,
				},
				encounterUpdater: &mockEncounterUpdater{
					retEnc: dummyEncounter2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.UpdateEncounter(tt.args.ctx, tt.args.userID, tt.args.encID, tt.args.enc),
				"UpdateEncounter(%v, %v, %v, %v)",
				tt.args.ctx,
				tt.args.userID,
				tt.args.encID,
				tt.args.enc,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_DeleteEncounter(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		encID  string
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		userID: dummyUser1.ID,
		encID:  dummyEncounter1.ID,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "delete encounter ok",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter1},
				encounterDeleter: &mockEncounterDeleter{},
			},
			args: dummyArgs,
			want: api.Result{
				Status: http.StatusOK, Name: "id", Value: dummyEncounter1.ID},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDeleter: &mockEncounterDeleter{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
				},
			},
		},
		{
			name: "delete encounter reader error",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter1, err: dummyError},
				encounterDeleter: &mockEncounterDeleter{},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
					err: dummyError,
				},
				encounterDeleter: &mockEncounterDeleter{},
			},
		},
		{
			name: "delete encounter leader check false",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter1},
				encounterDeleter: &mockEncounterDeleter{},
			},
			args: args{
				ctx:    dummyCtx,
				userID: dummyID,
				encID:  dummyEncounter1.ID,
			},
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDeleter: &mockEncounterDeleter{},
			},
		},
		{
			name: "delete encounter deleter error",
			mocks: mocks{
				encounterReader:  &mockEncounterReader{enc: dummyEncounter1},
				encounterDeleter: &mockEncounterDeleter{err: dummyError},
			},
			args: dummyArgs,
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDeleter: &mockEncounterDeleter{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					err: dummyError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.DeleteEncounter(tt.args.ctx, tt.args.userID, tt.args.encID),
				"DeleteEncounter(%v, %v, %v)",
				tt.args.ctx,
				tt.args.userID,
				tt.args.encID,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_ConfirmEncounter(t *testing.T) {
	type args struct {
		ctx    context.Context
		encID  string
		userID string
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		encID:  dummyEncounter1.ID,
		userID: dummyUser1.ID,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "confirm encounter ok",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:    &mockEncounterReader{enc: dummyEncounter1},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
			want: api.Result{Status: http.StatusOK, Name: "id", Value: dummyEncounter1.ID},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterConfirmer: &mockEncounterConfirmer{
					ctx:   dummyCtx,
					encID: dummyEncounter1.ID,
					user:  dummyUser1,
				},
			},
		},
		{
			name: "confirm encounter user not invited",
			args: args{
				ctx:    dummyArgs.ctx,
				encID:  dummyArgs.encID,
				userID: dummyID,
			},
			mocks: mocks{
				encounterReader:    &mockEncounterReader{enc: dummyEncounter1},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
		},
		{
			name: "confirm encounter user already confirmed",
			args: args{
				ctx:    dummyArgs.ctx,
				encID:  dummyArgs.encID,
				userID: dummyUser2.ID,
			},
			mocks: mocks{
				encounterReader:    &mockEncounterReader{enc: dummyEncounter1},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
			want: api.Result{Status: http.StatusUnprocessableEntity, Name: "error", Value: "user is already confirmed"},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
		},
		{
			name: "confirm encounter confirmer error",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:    &mockEncounterReader{enc: dummyEncounter1},
				encounterConfirmer: &mockEncounterConfirmer{err: dummyError},
			},
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterConfirmer: &mockEncounterConfirmer{
					ctx:   dummyCtx,
					encID: dummyEncounter1.ID,
					user:  dummyUser1,
					err:   dummyError,
				},
			},
		},
		{
			name: "confirm encounter reader error",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:    &mockEncounterReader{enc: dummyEncounter1, err: dummyError},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
					err: dummyError,
				},
				encounterConfirmer: &mockEncounterConfirmer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.ConfirmEncounter(tt.args.ctx, tt.args.userID, tt.args.encID),
				"ConfirmEncounter(%v, %v, %v)",
				tt.args.ctx,
				tt.args.userID,
				tt.args.encID,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

func TestAPI_DeclineEncounter(t *testing.T) {
	type args struct {
		ctx    context.Context
		encID  string
		userID string
	}
	dummyArgs := args{
		ctx:    dummyCtx,
		encID:  dummyEncounter1.ID,
		userID: dummyUser2.ID,
	}
	tests := []struct {
		name      string
		mocks     mocks
		args      args
		want      api.Result
		wantMocks mocks
	}{
		{
			name: "decline encounter ok",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:   &mockEncounterReader{enc: dummyEncounter1},
				encounterDecliner: &mockEncounterDecliner{},
			},
			want: api.Result{Status: http.StatusOK, Name: "id", Value: dummyEncounter1.ID},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDecliner: &mockEncounterDecliner{
					ctx:    dummyCtx,
					encID:  dummyEncounter1.ID,
					userID: dummyUser2.ID,
				},
			},
		},
		{
			name: "decline encounter user not invited",
			args: args{
				ctx:    dummyArgs.ctx,
				encID:  dummyArgs.encID,
				userID: dummyID,
			},
			mocks: mocks{
				encounterReader:   &mockEncounterReader{enc: dummyEncounter1},
				encounterDecliner: &mockEncounterDecliner{},
			},
			want: unauthorizedAPIError,
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDecliner: &mockEncounterDecliner{},
			},
		},
		{
			name: "decline encounter user not confirmed",
			args: args{
				ctx:    dummyArgs.ctx,
				encID:  dummyArgs.encID,
				userID: dummyUser1.ID,
			},
			mocks: mocks{
				encounterReader:   &mockEncounterReader{enc: dummyEncounter1},
				encounterDecliner: &mockEncounterDecliner{},
			},
			want: api.Result{Status: http.StatusUnprocessableEntity, Name: "error", Value: "user is not confirmed"},
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDecliner: &mockEncounterDecliner{},
			},
		},
		{
			name: "decline encounter reader error",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:   &mockEncounterReader{enc: dummyEncounter1, err: dummyError},
				encounterDecliner: &mockEncounterDecliner{},
			},
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
					err: dummyError,
				},
				encounterDecliner: &mockEncounterDecliner{},
			},
		},
		{
			name: "decline encounter decliner error",
			args: dummyArgs,
			mocks: mocks{
				encounterReader:   &mockEncounterReader{enc: dummyEncounter1},
				encounterDecliner: &mockEncounterDecliner{err: dummyError},
			},
			want: dummyAPIError(http.StatusNotFound),
			wantMocks: mocks{
				encounterReader: &mockEncounterReader{
					ctx: dummyCtx,
					id:  dummyEncounter1.ID,
					enc: dummyEncounter1,
				},
				encounterDecliner: &mockEncounterDecliner{
					ctx:    dummyCtx,
					encID:  dummyEncounter1.ID,
					userID: dummyUser2.ID,
					err:    dummyError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := apiFromMocks(tt.mocks)
			assert.Equalf(
				t,
				tt.want,
				a.DeclineEncounter(tt.args.ctx, tt.args.userID, tt.args.encID),
				"DeclineEncounter(%v, %v, %v)",
				tt.args.ctx,
				tt.args.userID,
				tt.args.encID,
			)
			assert.Equal(t, tt.wantMocks, tt.mocks)
		})
	}
}

type mockUserEncountersReader struct {
	ctx    context.Context
	userID string
	encs   []types.Encounter
	err    error
}

func (m *mockUserEncountersReader) ReadUserEncounters(ctx context.Context, userID string) ([]types.Encounter, error) {
	m.ctx = ctx
	m.userID = userID
	return m.encs, m.err
}

type mockEncounterReader struct {
	ctx context.Context
	id  string
	enc types.Encounter
	err error
}

func (m *mockEncounterReader) ReadEncounter(ctx context.Context, id string) (types.Encounter, error) {
	m.ctx = ctx
	m.id = id
	return m.enc, m.err
}

type mockEncounterCreator struct {
	ctx context.Context
	enc types.Encounter
	id  string
	err error
}

func (m *mockEncounterCreator) CreateEncounter(ctx context.Context, e types.Encounter) (string, error) {
	m.ctx = ctx
	m.enc = e
	return m.id, m.err
}

type mockLeaderChecker struct {
	ctx      context.Context
	token    string
	groupID  string
	isLeader bool
	err      error
}

func (m *mockLeaderChecker) IsGroupLeader(ctx context.Context, token string, groupID string) (bool, error) {
	m.ctx = ctx
	m.token = token
	m.groupID = groupID
	return m.isLeader, m.err
}

type mockEncounterUpdater struct {
	ctx    context.Context
	rcvEnc types.Encounter
	retEnc types.Encounter
	err    error
}

func (m *mockEncounterUpdater) UpdateEncounter(ctx context.Context, e types.Encounter) (types.Encounter, error) {
	m.ctx = ctx
	m.rcvEnc = e
	return m.retEnc, m.err
}

type mockEncounterDeleter struct {
	ctx context.Context
	id  string
	err error
}

func (m *mockEncounterDeleter) DeleteEncounter(ctx context.Context, id string) error {
	m.ctx = ctx
	m.id = id
	return m.err
}

type mockEncounterConfirmer struct {
	ctx   context.Context
	encID string
	user  types.User
	err   error
}

func (m *mockEncounterConfirmer) ConfirmEncounter(ctx context.Context, encID string, user types.User) error {
	m.ctx = ctx
	m.encID = encID
	m.user = user
	return m.err
}

type mockEncounterDecliner struct {
	ctx    context.Context
	encID  string
	userID string
	err    error
}

func (m *mockEncounterDecliner) DeclineEncounter(ctx context.Context, encID, userID string) error {
	m.ctx = ctx
	m.encID = encID
	m.userID = userID
	return m.err
}
