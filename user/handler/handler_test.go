package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gabrielseibel1/gaef-user-service/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func TestHandler_JWTAuthMiddleware_OK(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	// {
	// 	"sub": "1234567890",
	// 	"name": "John Doe",
	// 	"iat": 1516239022
	// }
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.5mhBHqs5_DTLdINd9p5m7ZJ6XD0Xc55kIaCRY5r6HRA"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	c.AddParam("id", "1234567890")

	// run code under test
	New(nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

	// assertions
	if got := w.Body.String(); got != "" {
		t.Errorf("got response body %s, want \"\"", got)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := c.GetString("AuthenticatedUserID"), "1234567890"; got != want {
		t.Errorf("got userID = %s, want %s", got, want)
	}
	if c.IsAborted() {
		t.Errorf("context was aborted")
	}
}

func TestHandler_JWTAuthMiddleware_NoJWT(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	c.Request = req
	c.AddParam("id", "1234567890")

	// run code under test
	New(nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing authorization header"; got != want {
		t.Errorf("got response body %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := c.GetString("AuthenticatedUserID"), ""; got != want {
		t.Errorf("got userID = %s, want %s", got, want)
	}
	if !c.IsAborted() {
		t.Errorf("context was not aborted")
	}
}

func TestHandler_JWTAuthMiddleware_InvalidJWTUser(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	// {
	// 	"sub": "0987654321",
	// 	"name": "John Doe",
	// 	"iat": 1516239022
	// }
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwOTg3NjU0MzIxIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.E1FM9mPvmrTd4GO-c8Zfnul77TFSBoUHmrZdYzlcf4w"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	c.AddParam("id", "1234567890")

	// run code under test
	New(nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "unauthorized"; got != want {
		t.Errorf("got response body %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := c.GetString("AuthenticatedUserID"), ""; got != want {
		t.Errorf("got userID = %s, want %s", got, want)
	}
	if !c.IsAborted() {
		t.Errorf("context was not aborted")
	}
}

func TestHandler_JWTAuthMiddleware_InvalidJWTSig(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	// {
	// 	"sub": "0987654321",
	// 	"name": "John Doe",
	// 	"iat": 1516239022
	// }
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwOTg3NjU0MzIxIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.E1FM9mPvmrTd4GO-c8Zfnul77TFSBoUHmrZdYzlcf4w"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	c.AddParam("id", "1234567890")

	// run code under test
	New(nil, nil, nil, nil, nil, []byte("not-the-one-used-to-sign-token")).JWTAuthMiddleware()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "invalid or expired token"; got != want {
		t.Errorf("got response body %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := c.GetString("AuthenticatedUserID"), ""; got != want {
		t.Errorf("got userID = %s, want %s", got, want)
	}
	if !c.IsAborted() {
		t.Errorf("context was not aborted")
	}
}

func TestHandler_JWTAuthMiddleware_EmptyJWTClaims(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := &http.Request{
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	// {}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.P4Lqll22jQQJ1eMJikvNg5HKG-cKB0hUZA9BZFIG7Jk"
	req.Header.Add("Authorization", "Bearer "+jwt)
	c.Request = req
	c.AddParam("id", "1234567890")

	// run code under test
	New(nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "invalid or expired token"; got != want {
		t.Errorf("got response body %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := c.GetString("AuthenticatedUserID"), ""; got != want {
		t.Errorf("got userID = %s, want %s", got, want)
	}
	if !c.IsAborted() {
		t.Errorf("context was not aborted")
	}
}

type mockCreator struct {
	// receive
	user     domain.User
	password string
	ctx      context.Context

	// return
	id  string
	err error
}

func (mc *mockCreator) Create(user *domain.User, password string, ctx context.Context) (string, error) {
	mc.user = *user
	mc.password = password
	mc.ctx = ctx
	return mc.id, mc.err
}

func TestHandler_Signup_OK(t *testing.T) {
	// prepare test setup
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "Gabriel do Souza Seibel",
		Email:    "gabriel.seibel@tuta.io",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(mockCreator, nil, nil, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.ID, mockCreator.id; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockCreator.user.Name, reqBody.Name; got != want {
		t.Errorf("got created name %s, want %s", got, want)
	}
	if got, want := mockCreator.user.Email, reqBody.Email; got != want {
		t.Errorf("got created email %s, want %s", got, want)
	}
	if got, want := mockCreator.password, reqBody.Password; got != want {
		t.Errorf("got created password %s, want %s", got, want)
	}
	if got, want := mockCreator.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
}

func TestHandler_Signup_MissingName(t *testing.T) {
	// prepare test setup
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "gabriel.seibel@tuta.io",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(mockCreator, nil, nil, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body error %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Signup_MissingEmail(t *testing.T) {
	// prepare test setup
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{
		Name:     "Gabriel do Souza Seibel",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(mockCreator, nil, nil, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body error %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Signup_MissingPassword(t *testing.T) {
	// prepare test setup
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Name:  "Gabriel do Souza Seibel",
		Email: "gabriel.seibel@tuta.io",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(mockCreator, nil, nil, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body error %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Signup_CreatorError(t *testing.T) {
	// prepare test setup
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: errors.New("mock creator error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "Gabriel do Souza Seibel",
		Email:    "gabriel.seibel@tuta.io",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(mockCreator, nil, nil, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "email is taken"; got != want {
		t.Errorf("got response body error %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

type mockLoginer struct {
	// received
	email    string
	password string
	ctx      context.Context

	// returned
	user *domain.User
	err  error
}

func (ml *mockLoginer) Login(email string, password string, ctx context.Context) (*domain.User, error) {
	ml.email = email
	ml.password = password
	ml.ctx = ctx
	return ml.user, ml.err
}

func TestHandler_Login_OK(t *testing.T) {
	// prepare test setup
	mockLoginer := &mockLoginer{
		user: &domain.User{
			ID:    "dummyID",
			Name:  "dummyName",
			Email: "dummyEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "gabriel.seibel@tuta.io",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req
	jwtSecret := []byte("test")

	// run code under test
	New(nil, mockLoginer, nil, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}

	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		t.Errorf("jwt.Parse() got error %s, want nil", err.Error())
	}
	if !token.Valid {
		t.Errorf("token is invalid: %s", token.Raw)
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	tokenName := claims["name"].(string)
	if got, want := tokenName, mockLoginer.user.Name; got != want {
		t.Errorf("got token claim \"name\": %s, want %s", got, want)
	}
	tokenEmail := claims["email"].(string)
	if got, want := tokenEmail, mockLoginer.user.Email; got != want {
		t.Errorf("got token claim \"email\": %s, want %s", got, want)
	}
	tokenSub := claims["sub"].(string)
	if got, want := tokenSub, mockLoginer.user.ID; got != want {
		t.Errorf("got token claim \"sub\": %s, want %s", got, want)
	}
	tokenExp := claims["exp"].(float64)
	wantExp := time.Now().Add(time.Hour * 24 * 7)
	gotExp := time.Unix(int64(tokenExp), 0)
	if wantExp.Sub(gotExp).Abs() > time.Second*1 {
		t.Errorf("got token claim \"exp\": %s, want 1s from %s", gotExp, wantExp)
	}

	if got, want := mockLoginer.email, reqBody.Email; got != want {
		t.Errorf("got loginer email = %s, want %s", got, want)
	}
	if got, want := mockLoginer.password, reqBody.Password; got != want {
		t.Errorf("got loginer password = %s, want %s", got, want)
	}
	if got, want := mockLoginer.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
}

func TestHandler_Login_MissingEmail(t *testing.T) {
	// prepare test setup
	mockLoginer := &mockLoginer{
		user: &domain.User{
			ID:    "dummyID",
			Name:  "dummyName",
			Email: "dummyEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Password string `json:"password"`
	}{
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req
	jwtSecret := []byte("test")

	// run code under test
	New(nil, mockLoginer, nil, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "unauthorized"; got != want {
		t.Errorf("got response body error: %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Login_MissingPassword(t *testing.T) {
	// prepare test setup
	mockLoginer := &mockLoginer{
		user: &domain.User{
			ID:    "dummyID",
			Name:  "dummyName",
			Email: "dummyEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Email string `json:"email"`
	}{
		Email: "gabriel.seibel@tuta.io",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req
	jwtSecret := []byte("test")

	// run code under test
	New(nil, mockLoginer, nil, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "unauthorized"; got != want {
		t.Errorf("got response body error: %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Login_LoginerError(t *testing.T) {
	// prepare test setup
	mockLoginer := &mockLoginer{
		user: &domain.User{
			ID:    "dummyID",
			Name:  "dummyName",
			Email: "dummyEmail",
		},
		err: errors.New("mock loginer error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "gabriel.seibel@tuta.io",
		Password: "test123",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req
	jwtSecret := []byte("test")

	// run code under test
	New(nil, mockLoginer, nil, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "unauthorized"; got != want {
		t.Errorf("got response body error: %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_GetIDFromToken(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, nil, nil, nil, nil).GetIDFromToken()(c)

	// assertions
	var resp struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.ID, dummyID; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

type mockReader struct {
	// receive
	id  string
	ctx context.Context

	// return
	user *domain.User
	err  error
}

func (mr *mockReader) Read(id string, ctx context.Context) (*domain.User, error) {
	mr.id = id
	mr.ctx = ctx
	return mr.user, mr.err
}

func TestHandler_GetUserFromID_OK(t *testing.T) {
	// prepare test setup
	mockReader := &mockReader{
		user: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, mockReader, nil, nil, nil).GetUserFromID()(c)

	// assertions
	var resp struct {
		User domain.User `json:"user"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.User, *mockReader.user; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockReader.id, dummyID; got != want {
		t.Errorf("mockReader.Read() received id %s, want %s", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Errorf("mockReader.Read() received id %v, want %v", got, want)
	}
}

func TestHandler_GetUserFromID_ReaderError(t *testing.T) {
	// prepare test setup
	mockReader := &mockReader{
		user: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: errors.New("mock reader error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, mockReader, nil, nil, nil).GetUserFromID()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "user not found"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockReader.id, dummyID; got != want {
		t.Errorf("mockReader.Read() received id %s, want %s", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Errorf("mockReader.Read() received id %v, want %v", got, want)
	}
}

type mockUpdater struct {
	// receive
	receiveUser *domain.User
	ctx         context.Context

	// return
	returnUser *domain.User
	err        error
}

func (mu *mockUpdater) Update(user *domain.User, ctx context.Context) (*domain.User, error) {
	mu.receiveUser = user
	mu.ctx = ctx
	return mu.returnUser, mu.err
}

func TestHandler_UpdateUser_OK(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		returnUser: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := domain.User{
		ID:    dummyID,
		Name:  "bodyName",
		Email: "bodyEmail",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		User domain.User `json:"user"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.User, *mockUpdater.returnUser; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := *mockUpdater.receiveUser, reqBody; got != want {
		t.Errorf("mockUpdater.Update() received user %s, want %s", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Errorf("mockUpdater.Update() received id %v, want %v", got, want)
	}
}

func TestHandler_UpdateUser_MismatchedIDs(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		returnUser: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("AuthenticatedUserID", "1")
	reqBody := domain.User{
		ID:    "2",
		Name:  "bodyName",
		Email: "bodyEmail",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "unauthorized"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_UpdateUser_UpdaterError(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		returnUser: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: errors.New("mock updater error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := domain.User{
		ID:    dummyID,
		Name:  "bodyName",
		Email: "bodyEmail",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "user not found"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := *mockUpdater.receiveUser, reqBody; got != want {
		t.Errorf("mockUpdater.Update() received user %s, want %s", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Errorf("mockUpdater.Update() received id %v, want %v", got, want)
	}
}

func TestHandler_UpdateUser_MissingUserName(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		returnUser: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: errors.New("mock updater error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := domain.User{
		ID:    dummyID,
		Email: "bodyEmail",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_UpdateUser_MissingUserEmail(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		returnUser: &domain.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: errors.New("mock updater error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := domain.User{
		ID:   dummyID,
		Name: "bodyName",
	}
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		t.Error("failed to marshal json")
	}
	req := &http.Request{
		Body: io.NopCloser(bytes.NewBufferString(string(reqBodyJson))),
	}
	c.Request = req

	// run code under test
	New(nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

type mockDeleter struct {
	// receive
	id  string
	ctx context.Context

	// return
	err error
}

func (md *mockDeleter) Delete(id string, ctx context.Context) error {
	md.id = id
	md.ctx = ctx
	return md.err
}

func TestHandler_Delete_OK(t *testing.T) {
	// prepare test setup
	mockDeleter := &mockDeleter{
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, nil, nil, mockDeleter, nil).DeleteUser()(c)

	// assertions
	var resp struct {
		Message string `json:"message"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Message, "deleted user "+dummyID; got != want {
		t.Errorf("got message %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockDeleter.id, dummyID; got != want {
		t.Errorf("mockDeleter.DeleteUser() received user %s, want %s", got, want)
	}
	if got, want := mockDeleter.ctx, c; got != want {
		t.Errorf("mockDeleter.DeleteUser() received id %v, want %v", got, want)
	}
}

func TestHandler_Delete_DeleterError(t *testing.T) {
	// prepare test setup
	mockDeleter := &mockDeleter{
		err: errors.New("mock deleter error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, nil, nil, mockDeleter, nil).DeleteUser()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "user not found"; got != want {
		t.Errorf("got message %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockDeleter.id, dummyID; got != want {
		t.Errorf("mockDeleter.DeleteUser() received user %s, want %s", got, want)
	}
	if got, want := mockDeleter.ctx, c; got != want {
		t.Errorf("mockDeleter.DeleteUser() received id %v, want %v", got, want)
	}
}
