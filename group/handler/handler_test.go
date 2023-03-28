package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gabrielseibel1/gaef/group/handler"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mocks struct {
	leaderChecker             handler.LeaderChecker
	groupCreator              handler.GroupCreator
	participatingGroupsReader handler.ParticipatingGroupsReader
	leadingGroupsReader       handler.LeadingGroupsReader
	groupReader               handler.GroupReader
	groupUpdater              handler.GroupUpdater
	groupDeleter              handler.GroupDeleter
}
type fields struct {
	mocks
	ctxParams map[string]string
	ctxValues map[string]any
	request   *http.Request
}
type test struct {
	name          string
	fields        fields
	responseOK    func(*httptest.ResponseRecorder) error
	sideEffectsOK func(mocks, *gin.Context) error
}

func testRequest(t *testing.T, tt test, cut func(handler.Handler) gin.HandlerFunc) {
	// setup test
	h := handler.New(
		tt.fields.leaderChecker,
		tt.fields.groupCreator,
		tt.fields.participatingGroupsReader,
		tt.fields.leadingGroupsReader,
		tt.fields.groupReader,
		tt.fields.groupUpdater,
		tt.fields.groupDeleter,
	)
	responseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(responseRecorder)
	for k, v := range tt.fields.ctxParams {
		c.AddParam(k, v)
	}
	for k, v := range tt.fields.ctxValues {
		c.Set(k, v)
	}
	c.Request = tt.fields.request

	// run code under test
	cut(h)(c)

	// assertions
	if err := tt.responseOK(responseRecorder); err != nil {
		t.Error(err)
	}
	if err := tt.sideEffectsOK(tt.fields.mocks, c); err != nil {
		t.Error(err)
	}
}

func TestHandler_CreateGroupHandler(t *testing.T) {
	tests := []test{
		{
			name: "create group ok",
			fields: fields{
				mocks: mocks{
					groupCreator: &mockGroupCreator{
						retGroup: dummyGroup2,
					},
				},
				request: &http.Request{Body: io.NopCloser(bytes.NewBuffer(dummyGroup1JSON))},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusCreated; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Group types.Group }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Group, dummyGroup2; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("resp.Group: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupCreator := m.groupCreator.(*mockGroupCreator)
				if got, want := groupCreator.ctx, ctx; got != want {
					return fmt.Errorf("groupCreator.ctx: got %v, want %v", got, want)
				}
				if got, want := groupCreator.rcvGroup, dummyGroup1; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupCreator.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "create group bad request",
			fields: fields{
				mocks: mocks{
					groupCreator: &mockGroupCreator{
						retGroup: dummyGroup2,
					},
				},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusBadRequest; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if resp.Error == "" {
					return errors.New("resp.Error: got empty error, want some")
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupCreator := m.groupCreator.(*mockGroupCreator)
				if got, want := groupCreator.ctx, nilCtx; got != want {
					return fmt.Errorf("groupCreator.ctx: got %v, want %v", got, want)
				}
				if got, want := groupCreator.rcvGroup, emptyGroup; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupCreator.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "create group creator error",
			fields: fields{
				mocks: mocks{
					groupCreator: &mockGroupCreator{
						retGroup: dummyGroup2,
						err:      dummyError,
					},
				},
				request: &http.Request{Body: io.NopCloser(bytes.NewBuffer(dummyGroup1JSON))},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusUnprocessableEntity; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupCreator := m.groupCreator.(*mockGroupCreator)
				if got, want := groupCreator.ctx, ctx; got != want {
					return fmt.Errorf("groupCreator.ctx: got %v, want %v", got, want)
				}
				if got, want := groupCreator.rcvGroup, dummyGroup1; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupCreator.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.CreateGroupHandler()
			})
		})
	}
}

func TestHandler_DeleteGroupHandler(t *testing.T) {
	tests := []test{
		{
			name: "delete group ok",
			fields: fields{
				mocks: mocks{
					groupDeleter: &mockGroupDeleter{},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusOK; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Message string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Message, "deleted group "+dummyGroup1.ID; got != want {
					return fmt.Errorf("resp.Message: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupDeleter := m.groupDeleter.(*mockGroupDeleter)
				if got, want := groupDeleter.ctx, ctx; got != want {
					return fmt.Errorf("groupDeleter.ctx: got %v, want %v", got, want)
				}
				if got, want := groupDeleter.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("groupDeleter.groupID: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "delete group deleter error",
			fields: fields{
				mocks: mocks{
					groupDeleter: &mockGroupDeleter{err: dummyError},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusNotFound; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupDeleter := m.groupDeleter.(*mockGroupDeleter)
				if got, want := groupDeleter.ctx, ctx; got != want {
					return fmt.Errorf("groupDeleter.ctx: got %v, want %v", got, want)
				}
				if got, want := groupDeleter.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("groupDeleter.groupID: got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.DeleteGroupHandler()
			})
		})
	}
}

func TestHandler_OnlyLeadersMiddleware(t *testing.T) {
	tests := []test{
		{
			name: "only leaders middleware ok",
			fields: fields{
				mocks: mocks{
					leaderChecker: &mockLeaderChecker{isLeader: true},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID}, // any ID is OK
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				leaderChecker := m.leaderChecker.(*mockLeaderChecker)
				if got, want := leaderChecker.ctx, ctx; got != want {
					return fmt.Errorf("leaderChecker.ctx got %v, want %v", got, want)
				}
				if got, want := leaderChecker.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("leaderChecker.userID got %v, want %v", got, want)
				}
				if got, want := leaderChecker.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("leaderChecker.groupID got %v, want %v", got, want)
				}
				if ctx.IsAborted() {
					return errors.New("context was aborted")
				}
				return nil
			},
		},
		{
			name: "only leaders middleware checker false",
			fields: fields{
				mocks: mocks{
					leaderChecker: &mockLeaderChecker{},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID}, // any ID is OK
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusUnauthorized; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, "unauthorized"; got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				leaderChecker := m.leaderChecker.(*mockLeaderChecker)
				if got, want := leaderChecker.ctx, ctx; got != want {
					return fmt.Errorf("leaderChecker.ctx got %v, want %v", got, want)
				}
				if got, want := leaderChecker.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("leaderChecker.userID got %v, want %v", got, want)
				}
				if got, want := leaderChecker.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("leaderChecker.groupID got %v, want %v", got, want)
				}
				if !ctx.IsAborted() {
					return errors.New("context was not aborted")
				}
				return nil
			},
		},
		{
			name: "only leaders middleware checker error",
			fields: fields{
				mocks: mocks{
					leaderChecker: &mockLeaderChecker{isLeader: true, err: dummyError},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID}, // any ID is OK
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusUnauthorized; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, "unauthorized"; got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				leaderChecker := m.leaderChecker.(*mockLeaderChecker)
				if got, want := leaderChecker.ctx, ctx; got != want {
					return fmt.Errorf("leaderChecker.ctx got %v, want %v", got, want)
				}
				if got, want := leaderChecker.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("leaderChecker.userID got %v, want %v", got, want)
				}
				if got, want := leaderChecker.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("leaderChecker.groupID got %v, want %v", got, want)
				}
				if !ctx.IsAborted() {
					return errors.New("context was not aborted")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.OnlyLeadersMiddleware()
			})
		})
	}
}

func TestHandler_ReadGroupHandler(t *testing.T) {
	tests := []test{
		{
			name: "read group ok",
			fields: fields{
				mocks: mocks{
					groupReader: &mockGroupReader{group: dummyGroup1},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusOK; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Group types.Group }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Group, dummyGroup1; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("resp.Group: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupReader := m.groupReader.(*mockGroupReader)
				if got, want := groupReader.ctx, ctx; got != want {
					return fmt.Errorf("groupReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupReader.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("groupReader.groupID got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "read group reader error",
			fields: fields{
				mocks: mocks{
					groupReader: &mockGroupReader{err: dummyError},
				},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusNotFound; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupReader := m.groupReader.(*mockGroupReader)
				if got, want := groupReader.ctx, ctx; got != want {
					return fmt.Errorf("groupReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupReader.groupID, dummyGroup1.ID; got != want {
					return fmt.Errorf("groupReader.groupID got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.ReadGroupHandler()
			})
		})
	}
}

func TestHandler_ReadLeadingGroupsHandler(t *testing.T) {
	tests := []test{
		{
			name: "read leading groups ok",
			fields: fields{
				mocks: mocks{
					leadingGroupsReader: &mockLeadingGroupsReader{
						groups: []types.Group{dummyGroup1, dummyGroup2},
					},
				},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusOK; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Groups []types.Group }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Groups, []types.Group{dummyGroup1, dummyGroup2}; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("resp.Group: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupsReader := m.leadingGroupsReader.(*mockLeadingGroupsReader)
				if got, want := groupsReader.ctx, ctx; got != want {
					return fmt.Errorf("groupsReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupsReader.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("groupsReader.userID got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "read leading groups reader error",
			fields: fields{
				mocks: mocks{
					leadingGroupsReader: &mockLeadingGroupsReader{err: dummyError},
				},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusNotFound; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupsReader := m.leadingGroupsReader.(*mockLeadingGroupsReader)
				if got, want := groupsReader.ctx, ctx; got != want {
					return fmt.Errorf("groupsReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupsReader.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("groupsReader.userID got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.ReadLeadingGroupsHandler()
			})
		})
	}
}

func TestHandler_ReadParticipatingGroupsHandler(t *testing.T) {
	tests := []test{
		{
			name: "read participating groups ok",
			fields: fields{
				mocks: mocks{
					participatingGroupsReader: &mockParticipatingGroupsReader{
						groups: []types.Group{dummyGroup1, dummyGroup2},
					},
				},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusOK; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Groups []types.Group }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Groups, []types.Group{dummyGroup1, dummyGroup2}; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("resp.Group: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupsReader := m.participatingGroupsReader.(*mockParticipatingGroupsReader)
				if got, want := groupsReader.ctx, ctx; got != want {
					return fmt.Errorf("groupsReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupsReader.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("groupsReader.userID got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "read leading groups reader error",
			fields: fields{
				mocks: mocks{
					participatingGroupsReader: &mockParticipatingGroupsReader{err: dummyError},
				},
				ctxValues: map[string]any{"userID": dummyGroup1.Members[0].ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusNotFound; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupsReader := m.participatingGroupsReader.(*mockParticipatingGroupsReader)
				if got, want := groupsReader.ctx, ctx; got != want {
					return fmt.Errorf("groupsReader.ctx got %v, want %v", got, want)
				}
				if got, want := groupsReader.userID, dummyGroup1.Members[0].ID; got != want {
					return fmt.Errorf("groupsReader.userID got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.ReadParticipatingGroupsHandler()
			})
		})
	}
}

type mockGroupUpdater struct {
	ctx      context.Context
	rcvGroup types.Group
	retGroup types.Group
	err      error
}

func (m *mockGroupUpdater) UpdateGroup(ctx context.Context, group types.Group) (types.Group, error) {
	m.ctx = ctx
	m.rcvGroup = group
	return m.retGroup, m.err
}

func TestHandler_UpdateGroupHandler(t *testing.T) {
	tests := []test{
		{
			name: "update group ok",
			fields: fields{
				mocks:     mocks{groupUpdater: &mockGroupUpdater{retGroup: dummyGroup2}},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
				request:   &http.Request{Body: io.NopCloser(bytes.NewBuffer(dummyGroup1JSON))},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusOK; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Group types.Group }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Group, dummyGroup2; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("resp.Group: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupUpdater := m.groupUpdater.(*mockGroupUpdater)
				if got, want := groupUpdater.ctx, ctx; got != want {
					return fmt.Errorf("groupUpdater.ctx: got %v, want %v", got, want)
				}
				if got, want := groupUpdater.rcvGroup, dummyGroup1; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupUpdater.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "update group bad request",
			fields: fields{
				mocks:     mocks{groupUpdater: &mockGroupUpdater{retGroup: dummyGroup2}},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusBadRequest; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if resp.Error == "" {
					return errors.New("resp.Error: got empty string, want some error")
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupUpdater := m.groupUpdater.(*mockGroupUpdater)
				if got, want := groupUpdater.ctx, nilCtx; got != want {
					return fmt.Errorf("groupUpdater.ctx: got %v, want %v", got, want)
				}
				if got, want := groupUpdater.rcvGroup, emptyGroup; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupUpdater.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "update group mismatching ids",
			fields: fields{
				mocks:     mocks{groupUpdater: &mockGroupUpdater{retGroup: dummyGroup2}},
				ctxParams: map[string]string{"id": dummyGroup2.ID}, // not the same as request body
				request:   &http.Request{Body: io.NopCloser(bytes.NewBuffer(dummyGroup1JSON))},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusUnprocessableEntity; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, "group id cannot be updated"; got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupUpdater := m.groupUpdater.(*mockGroupUpdater)
				if got, want := groupUpdater.ctx, nilCtx; got != want {
					return fmt.Errorf("groupUpdater.ctx: got %v, want %v", got, want)
				}
				if got, want := groupUpdater.rcvGroup, emptyGroup; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupUpdater.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
		{
			name: "update group updater error",
			fields: fields{
				mocks:     mocks{groupUpdater: &mockGroupUpdater{retGroup: dummyGroup2, err: dummyError}},
				ctxParams: map[string]string{"id": dummyGroup1.ID},
				request:   &http.Request{Body: io.NopCloser(bytes.NewBuffer(dummyGroup1JSON))},
			},
			responseOK: func(recorder *httptest.ResponseRecorder) error {
				if got, want := recorder.Result().StatusCode, http.StatusNotFound; got != want {
					return fmt.Errorf("recorder.Result().StatusCode: got %v, want %v", got, want)
				}
				var resp struct{ Error string }
				if err := json.NewDecoder(recorder.Result().Body).Decode(&resp); err != nil {
					return err
				}
				if got, want := resp.Error, dummyError.Error(); got != want {
					return fmt.Errorf("resp.Error: got %v, want %v", got, want)
				}
				return nil
			},
			sideEffectsOK: func(m mocks, ctx *gin.Context) error {
				groupUpdater := m.groupUpdater.(*mockGroupUpdater)
				if got, want := groupUpdater.ctx, ctx; got != want {
					return fmt.Errorf("groupUpdater.ctx: got %v, want %v", got, want)
				}
				if got, want := groupUpdater.rcvGroup, dummyGroup1; !reflect.DeepEqual(got, want) {
					return fmt.Errorf("groupUpdater.rcvGroup: got %v, want %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt, func(h handler.Handler) gin.HandlerFunc {
				return h.UpdateGroupHandler()
			})
		})
	}
}

// mocks

type mockGroupCreator struct {
	ctx      context.Context
	rcvGroup types.Group
	retGroup types.Group
	err      error
}

func (m *mockGroupCreator) CreateGroup(ctx context.Context, group types.Group) (types.Group, error) {
	m.ctx = ctx
	m.rcvGroup = group
	return m.retGroup, m.err
}

type mockGroupDeleter struct {
	ctx     context.Context
	groupID string
	err     error
}

func (m *mockGroupDeleter) DeleteGroup(ctx context.Context, id string) error {
	m.ctx = ctx
	m.groupID = id
	return m.err
}

type mockLeaderChecker struct {
	ctx      context.Context
	userID   string
	groupID  string
	isLeader bool
	err      error
}

func (m *mockLeaderChecker) IsLeader(ctx context.Context, userID string, groupID string) (bool, error) {
	m.ctx = ctx
	m.userID = userID
	m.groupID = groupID
	return m.isLeader, m.err
}

type mockGroupReader struct {
	ctx     context.Context
	groupID string
	group   types.Group
	err     error
}

func (m *mockGroupReader) ReadGroup(ctx context.Context, id string) (types.Group, error) {
	m.ctx = ctx
	m.groupID = id
	return m.group, m.err
}

type mockLeadingGroupsReader struct {
	ctx    context.Context
	userID string
	groups []types.Group
	err    error
}

func (m *mockLeadingGroupsReader) ReadLeadingGroups(ctx context.Context, userID string) ([]types.Group, error) {
	m.ctx = ctx
	m.userID = userID
	return m.groups, m.err
}

type mockParticipatingGroupsReader struct {
	ctx    context.Context
	userID string
	groups []types.Group
	err    error
}

func (m *mockParticipatingGroupsReader) ReadParticipatingGroups(ctx context.Context, userID string) ([]types.Group, error) {
	m.ctx = ctx
	m.userID = userID
	return m.groups, m.err
}

// empty values
var (
	nilCtx     context.Context = nil
	emptyGroup                 = types.Group{}
)

// dummies
var (
	dummyGroup1 = types.Group{
		ID:          "dummy-id-1",
		Name:        "dummy-name-1",
		PictureURL:  "example1.com",
		Description: "dummy-description-1",
		Members:     []types.User{{ID: "dummy-user-id-1-1", Name: "dummy-user-name-1-1"}},
		Leaders:     []types.User{{ID: "dummy-user-id-2-1", Name: "dummy-user-name-2-1"}},
	}
	dummyGroup2 = types.Group{
		ID:          "dummy-id-2",
		Name:        "dummy-name-2",
		PictureURL:  "example2.com",
		Description: "dummy-description-2",
		Members:     []types.User{{ID: "dummy-user-id-1-2", Name: "dummy-user-name-1-2"}},
		Leaders:     []types.User{{ID: "dummy-user-id-2-2", Name: "dummy-user-name-2-2"}},
	}
	dummyGroup1JSON, _ = json.Marshal(dummyGroup1)
	dummyError         = errors.New("dummy error")
)
