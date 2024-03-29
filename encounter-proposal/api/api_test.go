package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gabrielseibel1/gaef/types"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gabrielseibel1/gaef/encounter-proposal/api"
	"github.com/gin-gonic/gin"
)

func TestAPI_EPCreatorGroupLeaderCheckerMiddleware_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	dummyUserID := "dummy-user-id"
	c.Set("userID", dummyUserID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			Creator:                types.Group{ID: "dummy-group-id", Leaders: []types.User{{ID: dummyUserID}}},
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPCreatorGroupLeaderCheckerMiddleware()(c)

	// assertions

	// verify response body
	if got := w.Body.String(); got != "" {
		t.Fatalf("got response body %s, want \"\"", got)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreatorGroupLeaderCheckerMiddleware_ReaderError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	dummyUserID := "dummy-user-id"
	c.Set("userID", dummyUserID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			Creator:                types.Group{ID: "dummy-group-id", Leaders: []types.User{{ID: dummyUserID}}},
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPCreatorGroupLeaderCheckerMiddleware()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "unauthorized"; got != want {
		t.Fatalf("got response body %s, want %s", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreatorGroupLeaderCheckerMiddleware_LeaderFalse(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	dummyUserID := "dummy-user-id"
	c.Set("userID", dummyUserID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			Creator:                types.Group{ID: "dummy-group-id", Leaders: []types.User{{ID: dummyUserID + "change"}}},
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPCreatorGroupLeaderCheckerMiddleware()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "unauthorized"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
		Applications:           []types.Application{},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}
	mockCreator := mockEPCreator{
		returnEP: types.EncounterProposal{
			ID:                     "mock-id",
			EncounterSpecification: types.EncounterSpecification{Name: "mock-name"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		&mockCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		ID string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.ID, mockCreator.returnEP.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusCreated; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyEP.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_BadRequest(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}
	mockCreator := mockEPCreator{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		&mockCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, wantNot := resp.Error, ""; got == wantNot {
		t.Fatalf("got %v, want not %v", got, wantNot)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mock received values
	if got, want := mockLeaderChecker.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_CreatorError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
		Applications:           []types.Application{},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}
	mockCreator := mockEPCreator{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock EP creator error"),
	}

	// run code under test

	api.New(
		&mockCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockCreator.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusConflict; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyEP.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_LeaderFalse(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
		Applications:           []types.Application{},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: false,
		err:      nil,
	}
	mockCreator := mockEPCreator{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock EP creator error"),
	}

	// run code under test

	api.New(
		&mockCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "unauthorized"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyEP.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_LeaderError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
		Applications:           []types.Application{},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      errors.New("mock checker error"),
	}
	mockCreator := mockEPCreator{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock EP creator error"),
	}

	// run code under test

	api.New(
		&mockCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "unauthorized"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyEP.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockCreator.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingAllHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("page", "42")
	// setup mocks
	mockReader := mockPagedEPsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingAllHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposals []types.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposals, mockReader.eps; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.page, 42; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingAllHandler_BadRequest(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("page", "not a number")
	// setup mocks
	mockReader := mockPagedEPsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingAllHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, wantNot := resp.Error, ""; got == wantNot {
		t.Fatalf("got %v, want not %v", got, wantNot)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingAllHandler_ReaderError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("page", "42")
	// setup mocks
	mockReader := mockPagedEPsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingAllHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.page, 42; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingByUserHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockReader := mockByGroupIDsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: nil,
	}
	mockLister := mockGroupLister{
		groups: []types.Group{{ID: "group-id-1"}, {ID: "group-id-2"}},
		err:    nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLister,
		nil,
	).EPReadingByUserHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposals []types.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposals, mockReader.eps; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLister.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLister.token, dummyToken; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.groupIDs, []string{"group-id-1", "group-id-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingByUserHandler_ListerError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLister := mockGroupLister{
		groups: []types.Group{{ID: "group-id-1"}, {ID: "group-id-2"}},
		err:    errors.New("mock lister error"),
	}
	mockReader := mockByGroupIDsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLister,
		nil,
	).EPReadingByUserHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockLister.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLister.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLister.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.groupIDs, nilStringSlice; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingByUserHandler_ReaderError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyToken := "dummy-token"
	c.Set("token", dummyToken)
	// setup mocks
	mockLister := mockGroupLister{
		groups: []types.Group{{ID: "group-id-1"}, {ID: "group-id-2"}},
		err:    nil,
	}
	mockReader := mockByGroupIDsReader{
		eps: []types.EncounterProposal{
			{EncounterSpecification: types.EncounterSpecification{Name: "test1"}},
			{EncounterSpecification: types.EncounterSpecification{Name: "test2"}},
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockLister,
		nil,
	).EPReadingByUserHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLister.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLister.token, dummyToken; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.groupIDs, []string{"group-id-1", "group-id-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingByIDHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyID := "dummy-id"
	c.AddParam("epid", dummyID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingByIDHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal types.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockReader.ep; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPReadingByIDHandler_ReaderError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyID := "dummy-id"
	c.AddParam("epid", dummyID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingByIDHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := types.EncounterProposal{
		ID:                     dummyEPID,
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep:  dummyEP,
		err: nil,
	}
	mockUpdater := mockEPUpdater{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		&mockUpdater,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal types.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockUpdater.returnEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_BadRequest(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep: types.EncounterProposal{
			ID:                     dummyEPID,
			EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
		},
		err: nil,
	}
	mockUpdater := mockEPUpdater{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		&mockUpdater,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if resp.Error == "" {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_MismatchingID(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := types.EncounterProposal{
		ID:                     dummyEPID,
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	c.AddParam("epid", "mismatching-id")
	// setup mocks
	mockReader := mockByIDEPReader{
		ep:  dummyEP,
		err: nil,
	}
	mockUpdater := mockEPUpdater{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		&mockUpdater,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "cannot update id"; got != want {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_ReaderError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := types.EncounterProposal{
		ID:                     dummyEPID,
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep:  dummyEP,
		err: errors.New("mock reader error"),
	}
	mockUpdater := mockEPUpdater{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		&mockUpdater,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if resp.Error == "" {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_UpdaterError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := types.EncounterProposal{
		ID:                     dummyEPID,
		EncounterSpecification: types.EncounterSpecification{Name: "dummy"},
	}
	epJSON, err := json.Marshal(dummyEP)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockReader := mockByIDEPReader{
		ep:  dummyEP,
		err: nil,
	}
	mockUpdater := mockEPUpdater{
		returnEP: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock"},
		},
		err: errors.New("mock updater error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockReader,
		&mockUpdater,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if resp.Error == "" {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockUpdater.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPDeletionHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockDeleter := mockEPDeleter{err: nil}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockDeleter,
		nil,
		nil,
		nil,
		nil,
	).EPDeletionHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Message string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Message, "deleted encounter proposal "+dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPDeletionHandler_DeleterError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockDeleter := mockEPDeleter{err: errors.New("mock deleter error")}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockDeleter,
		nil,
		nil,
		nil,
		nil,
	).EPDeletionHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if resp.Error == "" {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := types.Application{
		Applicant: types.Group{ID: dummyGroupID},
	}
	epJSON, err := json.Marshal(dummyApp)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyUserID := "dummy-user-id"
	c.Set("token", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppAppender{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock appender proposal"},
		},
		err: nil,
	}
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockAppender,
		nil,
		nil,
		&mockLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Message string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Message, "applied for encounter proposal "+dummyEPID; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, dummyApp; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_BadRequest(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	dummyUserID := "dummy-user-id"
	c.Set("token", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppAppender{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock appender proposal"},
		},
		err: nil,
	}
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockAppender,
		nil,
		nil,
		&mockLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if resp.Error == "" {
		t.Fatalf("got response body error \"\", want some error")
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_LeaderCheckerFalse(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := types.Application{
		Applicant: types.Group{ID: dummyGroupID},
	}
	epJSON, err := json.Marshal(dummyApp)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyUserID := "dummy-user-id"
	c.Set("token", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppAppender{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock appender proposal"},
		},
		err: nil,
	}
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: false,
		err:      nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockAppender,
		nil,
		nil,
		&mockLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, "user is not a leader of applicant group"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_LeaderCheckerError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := types.Application{
		Applicant: types.Group{ID: dummyGroupID},
	}
	epJSON, err := json.Marshal(dummyApp)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyUserID := "dummy-user-id"
	c.Set("token", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppAppender{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock appender proposal"},
		},
		err: nil,
	}
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      errors.New("mock leader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockAppender,
		nil,
		nil,
		&mockLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockLeaderChecker.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_AppenderError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := types.Application{
		Applicant: types.Group{ID: dummyGroupID},
	}
	epJSON, err := json.Marshal(dummyApp)
	if err != nil {
		t.Fatalf("unable to marshal request body")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(epJSON)),
	}
	c.Request = req
	dummyUserID := "dummy-user-id"
	c.Set("token", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppAppender{
		ep: types.EncounterProposal{
			EncounterSpecification: types.EncounterSpecification{Name: "mock appender proposal"},
		},
		err: errors.New("mock appender error"),
	}
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockAppender,
		nil,
		nil,
		&mockLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockAppender.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.token, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, dummyApp; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppDeletionHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}
	dummyUserID := "dummy-user-id"
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	dummyAppID := "dummy-app-id"
	c.AddParam("appid", dummyAppID)

	// setup mocks
	mockDeleter := &mockAppDeleter{}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		mockDeleter,
		nil,
		nil,
	).AppDeletionHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Message string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Message, fmt.Sprintf("deleted application %s of encounter proposal %s", dummyAppID, dummyEPID); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.appID, dummyAppID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppDeletionHandler_DeleterError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}
	dummyUserID := "dummy-user-id"
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	dummyAppID := "dummy-app-id"
	c.AddParam("appid", dummyAppID)

	// setup mocks
	mockDeleter := &mockAppDeleter{err: errors.New("mock app deleter error")}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		mockDeleter,
		nil,
		nil,
	).AppDeletionHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockDeleter.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.appID, dummyAppID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockDeleter.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

type mockGroupLeaderChecker struct {
	// receive
	ctx     context.Context
	groupID string
	token   string

	// return
	isLeader bool
	err      error
}

func (m *mockGroupLeaderChecker) IsGroupLeader(ctx context.Context, token string, groupID string) (bool, error) {
	m.ctx = ctx
	m.groupID = groupID
	m.token = token
	return m.isLeader, m.err
}

type mockEPCreator struct {
	// receive
	ctx       context.Context
	receiveEP types.EncounterProposal

	// return
	returnEP types.EncounterProposal
	err      error
}

func (m *mockEPCreator) Create(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error) {
	m.ctx = ctx
	m.receiveEP = ep
	return m.returnEP, m.err
}

type mockPagedEPsReader struct {
	// receive
	ctx  context.Context
	page int

	// return
	eps []types.EncounterProposal
	err error
}

func (m *mockPagedEPsReader) ReadPaged(ctx context.Context, page int) ([]types.EncounterProposal, error) {
	m.ctx = ctx
	m.page = page
	return m.eps, m.err
}

type mockByGroupIDsReader struct {
	// receive
	ctx      context.Context
	groupIDs []string

	// return
	eps []types.EncounterProposal
	err error
}

func (m *mockByGroupIDsReader) ReadByGroupIDs(ctx context.Context, groupIDs []string) ([]types.EncounterProposal, error) {
	m.ctx = ctx
	m.groupIDs = groupIDs
	return m.eps, m.err
}

type mockGroupLister struct {
	// receive
	ctx   context.Context
	token string

	// return
	groups []types.Group
	err    error
}

func (m *mockGroupLister) LeadingGroups(ctx context.Context, token string) ([]types.Group, error) {
	m.ctx = ctx
	m.token = token
	return m.groups, m.err
}

type mockByIDEPReader struct {
	// receive
	ctx context.Context
	id  string

	// return
	ep  types.EncounterProposal
	err error
}

func (m *mockByIDEPReader) ReadByID(ctx context.Context, id string) (types.EncounterProposal, error) {
	m.ctx = ctx
	m.id = id
	return m.ep, m.err
}

type mockEPUpdater struct {
	// receive
	ctx       context.Context
	receiveEP types.EncounterProposal

	// return
	returnEP types.EncounterProposal
	err      error
}

func (m *mockEPUpdater) Update(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error) {
	m.ctx = ctx
	m.receiveEP = ep
	return m.returnEP, m.err
}

type mockEPDeleter struct {
	// receive
	ctx context.Context
	id  string

	// return
	err error
}

func (m *mockEPDeleter) Delete(ctx context.Context, id string) error {
	m.ctx = ctx
	m.id = id
	return m.err
}

type mockAppAppender struct {
	// receive
	ctx  context.Context
	epID string
	app  types.Application

	// return
	ep  types.EncounterProposal
	err error
}

func (m *mockAppAppender) AppendApplication(ctx context.Context, epID string, app types.Application) error {
	m.ctx = ctx
	m.epID = epID
	m.app = app
	return m.err
}

type mockAppDeleter struct {
	// receive
	ctx   context.Context
	epID  string
	appID string

	// return
	err error
}

func (m *mockAppDeleter) DeleteApplication(ctx context.Context, epID string, appID string) error {
	m.ctx = ctx
	m.epID = epID
	m.appID = appID
	return m.err
}

var (
	emptyEP                        = types.EncounterProposal{}
	emptyApp                       = types.Application{}
	nilStringSlice []string        = nil
	nilCtx         context.Context = nil
)
