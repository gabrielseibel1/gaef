package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gabrielseibel1/gaef/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
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

	auth.NewMiddlewareGenerator(
		&mockAuthenticator,
		"userID",
		"token",
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
	if got, want := c.GetString("token"), mockAuthenticator.token; got != want {
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

	auth.NewMiddlewareGenerator(
		&mockAuthenticator,
		"userID",
		"token",
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
	if got, want := c.GetString("token"), ""; got != want {
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

	auth.NewMiddlewareGenerator(
		&mockAuthenticator,
		"userID",
		"token",
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
	if got, want := c.GetString("token"), ""; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	// verify stopped handler chain
	if !c.IsAborted() {
		t.Fatalf("context was not aborted")
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

func (ma *mockAuthenticatedUserIDGetter) ReadToken(ctx context.Context, token string) (string, error) {
	ma.ctx = ctx
	ma.token = token
	return ma.id, ma.err
}
