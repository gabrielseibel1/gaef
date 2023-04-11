package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gabrielseibel1/gaef/encounter/server"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type serverMocks struct {
	encounterCreator      server.EncounterCreator
	encounterReaderByUser server.EncounterReaderByUser
	encounterReaderByID   server.EncounterReaderByID
	encounterUpdater      server.EncounterUpdater
	encounterDeleter      server.EncounterDeleter
	encounterConfirmer    server.EncounterConfirmer
	encounterDecliner     server.EncounterDecliner
}

func fromMocks(m serverMocks) server.Server {
	return server.New(
		m.encounterCreator,
		m.encounterReaderByUser,
		m.encounterReaderByID,
		m.encounterUpdater,
		m.encounterDeleter,
		m.encounterConfirmer,
		m.encounterDecliner,
	)
}

type responseOKAsserter func(t *testing.T, recorder *httptest.ResponseRecorder)

type mocksOKAsserter func(t *testing.T, c context.Context, mocks serverMocks)

type test struct {
	name             string
	mocks            serverMocks
	request          *http.Request
	ctxParams        map[string]string
	ctxValues        map[string]any
	codeUnderTest    func(server.Server) gin.HandlerFunc
	assertResponseOK responseOKAsserter
	assertMocksOK    mocksOKAsserter
}

func testRequest(t *testing.T, tt test) {
	// setup test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = tt.request
	for k, v := range tt.ctxParams {
		c.AddParam(k, v)
	}
	for k, v := range tt.ctxValues {
		c.Set(k, v)
	}

	// run code under test
	s := fromMocks(tt.mocks)
	h := tt.codeUnderTest(s)
	h(c)

	// verify assertions
	tt.assertMocksOK(t, c, tt.mocks)
	tt.assertResponseOK(t, w)
}

func TestServer_CreateEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "create encounter handler ok",
			mocks: serverMocks{
				encounterCreator: &mockEncounterCreator{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"token": dummyToken},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.CreateEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterCreator, &mockEncounterCreator{
					ctx:   c,
					token: dummyToken,
					enc:   dummyEncounter1,
					res:   dummyResult,
				})
			},
		},
		{
			name: "create encounter handler bad request",
			mocks: serverMocks{
				encounterCreator: &mockEncounterCreator{res: dummyResult},
			},
			ctxValues:        map[string]any{"token": dummyToken},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.CreateEncounterHandler() },
			assertResponseOK: assertBodyFromError,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterCreator, &mockEncounterCreator{
					res: dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_ReadUserEncountersHandler(t *testing.T) {
	tests := []test{
		{
			name: "read encounter by user handler ok",
			mocks: serverMocks{
				encounterReaderByUser: &mockEncounterReaderByUser{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.ReadUserEncountersHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterReaderByUser, &mockEncounterReaderByUser{
					ctx:    c,
					userID: dummyUser1.ID,
					res:    dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_ReadEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "read encounter by id handler ok",
			mocks: serverMocks{
				encounterReaderByID: &mockEncounterReaderByID{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.ReadEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterReaderByID, &mockEncounterReaderByID{
					ctx:    c,
					userID: dummyUser1.ID,
					encID:  dummyEncounter1.ID,
					res:    dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_UpdateEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "update encounter handler ok",
			mocks: serverMocks{
				encounterUpdater: &mockEncounterUpdater{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.UpdateEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterUpdater, &mockEncounterUpdater{
					ctx:    c,
					userID: dummyUser1.ID,
					encID:  dummyEncounter1.ID,
					enc:    dummyEncounter1,
					res:    dummyResult,
				})
			},
		},
		{
			name: "update encounter handler bad request",
			mocks: serverMocks{
				encounterUpdater: &mockEncounterUpdater{res: dummyResult},
			},
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.UpdateEncounterHandler() },
			assertResponseOK: assertBodyFromError,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterUpdater, &mockEncounterUpdater{res: dummyResult})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_DeleteEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "delete encounter handler ok",
			mocks: serverMocks{
				encounterDeleter: &mockEncounterDeleter{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.DeleteEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterDeleter, &mockEncounterDeleter{
					ctx:    c,
					userID: dummyUser1.ID,
					encID:  dummyEncounter1.ID,
					res:    dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_ConfirmEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "confirm encounter handler ok",
			mocks: serverMocks{
				encounterConfirmer: &mockEncounterConfirmer{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.ConfirmEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterConfirmer, &mockEncounterConfirmer{
					ctx:    c,
					userID: dummyUser1.ID,
					encID:  dummyEncounter1.ID,
					res:    dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

func TestServer_DeclineEncounterHandler(t *testing.T) {
	tests := []test{
		{
			name: "decline encounter handler ok",
			mocks: serverMocks{
				encounterDecliner: &mockEncounterDecliner{res: dummyResult},
			},
			request:          requestWithEncounterInBody(t, dummyEncounter1),
			ctxValues:        map[string]any{"userID": dummyUser1.ID},
			ctxParams:        map[string]string{"encounter-id": dummyEncounter1.ID},
			codeUnderTest:    func(s server.Server) gin.HandlerFunc { return s.DeclineEncounterHandler() },
			assertResponseOK: assertBodyFromDummyResult,
			assertMocksOK: func(t *testing.T, c context.Context, mocks serverMocks) {
				assert.Equal(t, mocks.encounterDecliner, &mockEncounterDecliner{
					ctx:    c,
					userID: dummyUser1.ID,
					encID:  dummyEncounter1.ID,
					res:    dummyResult,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, tt)
		})
	}
}

var (
	dummyResult = result{
		s: 299, // this value interferes in how ctx.JSON processes the response *body* too, not only the status code
		k: "dummyKey",
		v: "dummyValue",
	}
	dummyToken = "dummy-token"
	dummyUser1 = types.User{
		ID:    "dummy-user-id-1",
		Name:  "dummy-user-name-1",
		Email: "dummy-user-email-1",
	}
	dummyGroup1 = types.Group{
		ID:          "dummy-group-id-1",
		Name:        "dummy-group-name-1",
		PictureURL:  "dummy-group-picture-url-1",
		Description: "dummy-group-description-1",
		Members:     []types.User{dummyUser1},
		Leaders:     []types.User{dummyUser1},
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
			Time: time.Now().Round(time.Minute).UTC(),
		},
		Groups:         []types.Group{dummyGroup1},
		InvitedUsers:   []types.User{dummyUser1},
		ConfirmedUsers: []types.User{dummyUser1},
	}
)

func requestWithEncounterInBody(t *testing.T, e types.Encounter) *http.Request {
	bodyBytes, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}
}

func assertBodyFromDummyResult(t *testing.T, r *httptest.ResponseRecorder) {
	assert.Equal(t, dummyResult.s, r.Result().StatusCode)
	var resp struct{ DummyKey string }
	err := json.NewDecoder(r.Result().Body).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, dummyResult.v, resp.DummyKey)
}

func assertBodyFromError(t *testing.T, r *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusBadRequest, r.Result().StatusCode)
	var resp struct{ Error string }
	if err := json.NewDecoder(r.Result().Body).Decode(&resp); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, resp.Error)
}

type result struct {
	s int
	k string
	v any
}

func (r result) S() int {
	return r.s
}

func (r result) K() string {
	return r.k
}

func (r result) V() any {
	return r.v
}

type mockEncounterCreator struct {
	ctx   context.Context
	token string
	enc   types.Encounter
	res   server.Result
}

func (m *mockEncounterCreator) CreateEncounter(ctx context.Context, token string, e types.Encounter) server.Result {
	m.ctx = ctx
	m.token = token
	m.enc = e
	return m.res
}

type mockEncounterReaderByUser struct {
	ctx    context.Context
	userID string
	res    server.Result
}

func (m *mockEncounterReaderByUser) ReadUserEncounters(ctx context.Context, userID string) server.Result {
	m.ctx = ctx
	m.userID = userID
	return m.res
}

type mockEncounterReaderByID struct {
	ctx    context.Context
	userID string
	encID  string
	res    server.Result
}

func (m *mockEncounterReaderByID) ReadEncounter(ctx context.Context, userID, encID string) server.Result {
	m.ctx = ctx
	m.userID = userID
	m.encID = encID
	return m.res
}

type mockEncounterUpdater struct {
	ctx    context.Context
	userID string
	encID  string
	enc    types.Encounter
	res    server.Result
}

func (m *mockEncounterUpdater) UpdateEncounter(ctx context.Context, userID string, encID string, e types.Encounter) server.Result {
	m.ctx = ctx
	m.userID = userID
	m.encID = encID
	m.enc = e
	return m.res
}

type mockEncounterDeleter struct {
	ctx    context.Context
	userID string
	encID  string
	res    server.Result
}

func (m *mockEncounterDeleter) DeleteEncounter(ctx context.Context, userID, encID string) server.Result {
	m.ctx = ctx
	m.userID = userID
	m.encID = encID
	return m.res
}

type mockEncounterConfirmer struct {
	ctx    context.Context
	userID string
	encID  string
	res    server.Result
}

func (m *mockEncounterConfirmer) ConfirmEncounter(ctx context.Context, userID, encID string) server.Result {
	m.ctx = ctx
	m.userID = userID
	m.encID = encID
	return m.res
}

type mockEncounterDecliner struct {
	ctx    context.Context
	userID string
	encID  string
	res    server.Result
}

func (m *mockEncounterDecliner) DeclineEncounter(ctx context.Context, userID, encID string) server.Result {
	m.ctx = ctx
	m.userID = userID
	m.encID = encID
	return m.res
}
