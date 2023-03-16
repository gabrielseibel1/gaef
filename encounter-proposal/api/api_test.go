package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gabrielseibel1/gaef-encounter-proposal-service/api"
	"github.com/gabrielseibel1/gaef-encounter-proposal-service/domain"
	"github.com/gin-gonic/gin"
)

func TestAPI_AuthMiddleware_OK(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Header: make(http.Header),
	}
	jwt := "test-token"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	// setup mocks
	mockAuthenticator := mockAuthenticatedUserIDGetter{
		id:  "test-user-id",
		err: nil,
	}

	// run code under test

	api.New(
		&mockAuthenticator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).AuthMiddleware()(c)

	// assertions

	// verify response body
	if got, want := w.Body.String(), ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := c.GetString("userID"), mockAuthenticator.id; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockAuthenticator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAuthenticator.token, jwt; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify not stopped handler chain
	if c.IsAborted() {
		t.Fatalf("context was aborted")
	}
}

func TestAPI_AuthMiddleware_AuthenticatorError(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		Header: make(http.Header),
	}
	jwt := "test-token"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	// setup mocks
	mockAuthenticator := mockAuthenticatedUserIDGetter{
		id:  "test-user-id",
		err: errors.New("mock authenticator error"),
	}

	// run code under test

	api.New(
		&mockAuthenticator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).AuthMiddleware()(c)

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
	if got, want := c.GetString("userID"), ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockAuthenticator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAuthenticator.token, jwt; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify stopped handler chain
	if !c.IsAborted() {
		t.Fatalf("context was not aborted")
	}
}

func TestAPI_AuthMiddleware_MissingHeader(t *testing.T) {
	// prepare test setup

	// setup request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{}
	c.Request = req
	// setup mocks
	mockAuthenticator := mockAuthenticatedUserIDGetter{
		id:  "test-user-id",
		err: nil,
	}

	// run code under test

	api.New(
		&mockAuthenticator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).AuthMiddleware()(c)

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
	if got, want := c.GetString("userID"), ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify stopped handler chain
	if !c.IsAborted() {
		t.Fatalf("context was not aborted")
	}
}

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
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Creator: domain.Group{ID: "dummy-group-id"},
			Name:    "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
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
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, mockByIDEPReader.ep.Creator.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.userID, dummyUserID; got != want {
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
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      nil,
	}
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Creator: domain.Group{ID: "dummy-group-id"},
			Name:    "mock",
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
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
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.userID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreatorGroupLeaderCheckerMiddleware_LeaderError(t *testing.T) {
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
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: true,
		err:      errors.New("mock leader error"),
	}
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Creator: domain.Group{ID: "dummy-group-id"},
			Name:    "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
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
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, mockByIDEPReader.ep.Creator.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.userID, dummyUserID; got != want {
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
	mockLeaderChecker := mockGroupLeaderChecker{
		isLeader: false,
		err:      nil,
	}
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Creator: domain.Group{ID: "dummy-group-id"},
			Name:    "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
		nil,
		nil,
		nil,
		&mockLeaderChecker,
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
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.groupID, mockByIDEPReader.ep.Creator.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockLeaderChecker.userID, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPCreationHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := domain.EncounterProposal{
		Name: "dummy",
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
	// setup mocks
	mockEPCreator := mockEPCreator{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		&mockEPCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockEPCreator.returnEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusCreated; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockEPCreator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPCreator.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
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
	// setup mocks
	mockEPCreator := mockEPCreator{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		&mockEPCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
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
}

func TestAPI_EPCreationHandler_CreatorError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEP := domain.EncounterProposal{
		Name: "dummy",
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
	// setup mocks
	mockEPCreator := mockEPCreator{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: errors.New("mock EP creator error"),
	}

	// run code under test

	api.New(
		nil,
		&mockEPCreator,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockEPCreator.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusConflict; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockEPCreator.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPCreator.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
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
	mockPagedEPsReader := mockPagedEPsReader{
		eps: []domain.EncounterProposal{{Name: "test1"}, {Name: "test2"}},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockPagedEPsReader,
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
		EncounterProposals []domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposals, mockPagedEPsReader.eps; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockPagedEPsReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockPagedEPsReader.page, 42; got != want {
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
	mockPagedEPsReader := mockPagedEPsReader{
		eps: []domain.EncounterProposal{{Name: "test1"}, {Name: "test2"}},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockPagedEPsReader,
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
	mockPagedEPsReader := mockPagedEPsReader{
		eps: []domain.EncounterProposal{{Name: "test1"}, {Name: "test2"}},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		&mockPagedEPsReader,
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
	if got, want := resp.Error, mockPagedEPsReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockPagedEPsReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockPagedEPsReader.page, 42; got != want {
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
	dummyID := "dummy-id"
	c.Set("userID", dummyID)
	// setup mocks
	mockByUserEPsReader := mockByUserEPsReader{
		eps: []domain.EncounterProposal{{Name: "test1"}, {Name: "test2"}},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockByUserEPsReader,
		nil,
		nil,
		nil,
		nil,
		nil,
	).EPReadingByUserHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposals []domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposals, mockByUserEPsReader.eps; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockByUserEPsReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByUserEPsReader.id, dummyID; got != want {
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
	dummyID := "dummy-id"
	c.Set("userID", dummyID)
	// setup mocks
	mockByUserEPsReader := mockByUserEPsReader{
		eps: []domain.EncounterProposal{{Name: "test1"}, {Name: "test2"}},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		&mockByUserEPsReader,
		nil,
		nil,
		nil,
		nil,
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
	if got, want := resp.Error, mockByUserEPsReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockByUserEPsReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByUserEPsReader.id, dummyID; got != want {
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
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
		nil,
		nil,
		nil,
		nil,
	).EPReadingByIDHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockByIDEPReader.ep; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyID; got != want {
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
	mockByIDEPReader := mockByIDEPReader{
		ep: domain.EncounterProposal{
			Name: "mock",
		},
		err: errors.New("mock reader error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		&mockByIDEPReader,
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
	if got, want := resp.Error, mockByIDEPReader.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockByIDEPReader.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockByIDEPReader.id, dummyID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := domain.EncounterProposal{
		ID:   dummyEPID,
		Name: "dummy",
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
	mockEPUpdater := mockEPUpdater{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPUpdater,
		nil,
		nil,
		nil,
	).EPUpdateHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockEPUpdater.returnEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockEPUpdater.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPUpdater.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
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
	mockEPUpdater := mockEPUpdater{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPUpdater,
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
	if got, want := mockEPUpdater.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPUpdater.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_MismatchingID(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := domain.EncounterProposal{
		ID:   dummyEPID,
		Name: "dummy",
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
	mockEPUpdater := mockEPUpdater{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: nil,
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPUpdater,
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
	if got, want := mockEPUpdater.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPUpdater.receiveEP, emptyEP; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_EPUpdateHandler_UpdaterError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyEPID := "dummy-ep-id"
	dummyEP := domain.EncounterProposal{
		ID:   dummyEPID,
		Name: "dummy",
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
	mockEPUpdater := mockEPUpdater{
		returnEP: domain.EncounterProposal{
			Name: "mock",
		},
		err: errors.New("mock updater error"),
	}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPUpdater,
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
	if got, want := mockEPUpdater.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPUpdater.receiveEP, dummyEP; !reflect.DeepEqual(got, want) {
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
	mockEPDeleter := mockEPDeleter{err: nil}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPDeleter,
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
	if got, want := mockEPDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPDeleter.id, dummyEPID; got != want {
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
	mockEPDeleter := mockEPDeleter{err: errors.New("mock deleter error")}

	// run code under test

	api.New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&mockEPDeleter,
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
	if got, want := mockEPDeleter.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockEPDeleter.id, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_OK(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := domain.Application{
		Creator: domain.Group{ID: dummyGroupID},
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
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppender{
		ep: domain.EncounterProposal{
			Name: "mock appender proposal",
		},
		err: nil,
	}
	mockGroupLeaderChecker := mockGroupLeaderChecker{
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
		nil,
		&mockAppender,
		&mockGroupLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		EncounterProposal domain.EncounterProposal
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.EncounterProposal, mockAppender.ep; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockGroupLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.userID, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, dummyApp; got != want {
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
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppender{
		ep: domain.EncounterProposal{
			Name: "mock appender proposal",
		},
		err: nil,
	}
	mockGroupLeaderChecker := mockGroupLeaderChecker{
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
		nil,
		&mockAppender,
		&mockGroupLeaderChecker,
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
	if got, want := mockGroupLeaderChecker.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.groupID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.userID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_LeaderCheckerFalse(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := domain.Application{
		Creator: domain.Group{ID: dummyGroupID},
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
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppender{
		ep: domain.EncounterProposal{
			Name: "mock appender proposal",
		},
		err: nil,
	}
	mockGroupLeaderChecker := mockGroupLeaderChecker{
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
		nil,
		&mockAppender,
		&mockGroupLeaderChecker,
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
	if got, want := mockGroupLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.userID, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_LeaderCheckerError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := domain.Application{
		Creator: domain.Group{ID: dummyGroupID},
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
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppender{
		ep: domain.EncounterProposal{
			Name: "mock appender proposal",
		},
		err: nil,
	}
	mockGroupLeaderChecker := mockGroupLeaderChecker{
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
		nil,
		&mockAppender,
		&mockGroupLeaderChecker,
	).AppCreationHandler()(c)

	// assertions

	// verify response body
	var resp struct {
		Error string
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("unable to decode response body to json")
	}
	if got, want := resp.Error, mockGroupLeaderChecker.err.Error(); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify response status code
	if got, want := w.Result().StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify mocks received values
	if got, want := mockGroupLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.userID, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, nilCtx; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, emptyApp; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAPI_AppCreationHandler_AppenderError(t *testing.T) {
	// prepare test setup

	// setup request
	dummyGroupID := "dummy-group-id"
	dummyApp := domain.Application{
		Creator: domain.Group{ID: dummyGroupID},
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
	c.Set("userID", dummyUserID)
	dummyEPID := "dummy-ep-id"
	c.AddParam("epid", dummyEPID)
	// setup mocks
	mockAppender := mockAppender{
		ep: domain.EncounterProposal{
			Name: "mock appender proposal",
		},
		err: errors.New("mock appender error"),
	}
	mockGroupLeaderChecker := mockGroupLeaderChecker{
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
		nil,
		&mockAppender,
		&mockGroupLeaderChecker,
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
	if got, want := mockGroupLeaderChecker.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.groupID, dummyGroupID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockGroupLeaderChecker.userID, dummyUserID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.ctx, c; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.epID, dummyEPID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := mockAppender.app, dummyApp; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

type mockAuthenticatedUserIDGetter struct {
	// receive
	ctx   context.Context
	token string

	//return
	id  string
	err error
}

func (ma *mockAuthenticatedUserIDGetter) GetAuthenticatedUserID(ctx context.Context, token string) (string, error) {
	ma.ctx = ctx
	ma.token = token
	return ma.id, ma.err
}

type mockGroupLeaderChecker struct {
	// receive
	ctx     context.Context
	groupID string
	userID  string

	// return
	isLeader bool
	err      error
}

func (m *mockGroupLeaderChecker) IsGroupLeader(ctx context.Context, groupID string, userID string) (bool, error) {
	m.ctx = ctx
	m.groupID = groupID
	m.userID = userID
	return m.isLeader, m.err
}

type mockEPCreator struct {
	// receive
	ctx       context.Context
	receiveEP domain.EncounterProposal

	// return
	returnEP domain.EncounterProposal
	err      error
}

func (m *mockEPCreator) Create(ctx context.Context, ep domain.EncounterProposal) (domain.EncounterProposal, error) {
	m.ctx = ctx
	m.receiveEP = ep
	return m.returnEP, m.err
}

type mockPagedEPsReader struct {
	// receive
	ctx  context.Context
	page int

	// return
	eps []domain.EncounterProposal
	err error
}

func (m *mockPagedEPsReader) ReadPaged(ctx context.Context, page int) ([]domain.EncounterProposal, error) {
	m.ctx = ctx
	m.page = page
	return m.eps, m.err
}

type mockByUserEPsReader struct {
	// receive
	ctx context.Context
	id  string

	// return
	eps []domain.EncounterProposal
	err error
}

func (m *mockByUserEPsReader) ReadByUser(ctx context.Context, id string) ([]domain.EncounterProposal, error) {
	m.ctx = ctx
	m.id = id
	return m.eps, m.err
}

type mockByIDEPReader struct {
	// receive
	ctx context.Context
	id  string

	// return
	ep  domain.EncounterProposal
	err error
}

func (m *mockByIDEPReader) ReadByID(ctx context.Context, id string) (domain.EncounterProposal, error) {
	m.ctx = ctx
	m.id = id
	return m.ep, m.err
}

type mockEPUpdater struct {
	// receive
	ctx       context.Context
	receiveEP domain.EncounterProposal

	// return
	returnEP domain.EncounterProposal
	err      error
}

func (m *mockEPUpdater) Update(ctx context.Context, ep domain.EncounterProposal) (domain.EncounterProposal, error) {
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

type mockAppender struct {
	// receive
	ctx  context.Context
	epID string
	app  domain.Application

	// return
	ep  domain.EncounterProposal
	err error
}

func (m *mockAppender) Append(ctx context.Context, epID string, app domain.Application) (domain.EncounterProposal, error) {
	m.ctx = ctx
	m.epID = epID
	m.app = app
	return m.ep, m.err
}

var (
	emptyEP  domain.EncounterProposal = domain.EncounterProposal{}
	emptyApp domain.Application       = domain.Application{}
	nilCtx   context.Context          = nil
)
