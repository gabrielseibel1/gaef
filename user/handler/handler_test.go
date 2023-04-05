package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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
	New(nil, nil, nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

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
	New(nil, nil, nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

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
	New(nil, nil, nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

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
	New(nil, nil, nil, nil, nil, nil, nil, []byte("not-the-one-used-to-sign-token")).JWTAuthMiddleware()(c)

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
	New(nil, nil, nil, nil, nil, nil, nil, []byte("test")).JWTAuthMiddleware()(c)

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
	user     types.User
	password string
	ctx      context.Context

	// return
	id  string
	err error
}

func (mc *mockCreator) Create(ctx context.Context, user types.UserWithHashedPassword) (string, error) {
	mc.user = user.User
	mc.password = user.HashedPassword
	mc.ctx = ctx
	return mc.id, mc.err
}

type mockByEmailReader struct {
	// receive
	ctx   context.Context
	email string

	// return
	user types.UserWithHashedPassword
	err  error
}

func (m *mockByEmailReader) ReadSensitiveByEmail(ctx context.Context, email string) (types.UserWithHashedPassword, error) {
	m.ctx = ctx
	m.email = email
	return m.user, m.err
}

type mockHasher struct {
	// receive
	password string

	// return
	hash string
	err  error
}

func (m *mockHasher) GenerateFromPassword(password string) (string, error) {
	m.password = password
	return m.hash, m.err
}

func TestHandler_Signup_OK(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		err: errors.New("mock email reader error"),
	}
	mockHasher := &mockHasher{
		hash: "dummy-hashed-password",
		err:  nil,
	}
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		User     types.User
		Password string
	}{
		User: types.User{
			Name:  "Gabriel do Souza Seibel",
			Email: "gabriel.seibel@tuta.io",
		},
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
	New(mockHasher, nil, mockCreator, nil, mockByEmailReader, nil, nil, nil).Signup()(c)

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
	if got, want := mockByEmailReader.email, reqBody.User.Email; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := mockHasher.password, reqBody.Password; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockCreator.user.Name, reqBody.User.Name; got != want {
		t.Errorf("got created name %s, want %s", got, want)
	}
	if got, want := mockCreator.user.Email, reqBody.User.Email; got != want {
		t.Errorf("got created email %s, want %s", got, want)
	}
	if got, want := mockCreator.password, mockHasher.hash; got != want {
		t.Errorf("got created password %s, want %s", got, want)
	}
	if got, want := mockCreator.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
}

func TestHandler_Signup_ReaderNilError(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		err: nil,
	}
	mockHasher := &mockHasher{
		hash: "dummy-hashed-password",
		err:  nil,
	}
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		User     types.User
		Password string
	}{
		User: types.User{
			Name:  "Gabriel do Souza Seibel",
			Email: "gabriel.seibel@tuta.io",
		},
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
	New(mockHasher, nil, mockCreator, nil, mockByEmailReader, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Error string
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Error, "email is taken"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnprocessableEntity; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.User.Email; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHandler_Signup_HasherError(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		err: errors.New("mock email reader error"),
	}
	mockHasher := &mockHasher{
		hash: "dummy-hashed-password",
		err:  errors.New("mock hasher error"),
	}
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		User     types.User
		Password string
	}{
		User: types.User{
			Name:  "Gabriel do Souza Seibel",
			Email: "gabriel.seibel@tuta.io",
		},
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
	New(mockHasher, nil, mockCreator, nil, mockByEmailReader, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Error string
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Error, "bad password"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnprocessableEntity; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.User.Email; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := mockHasher.password, reqBody.Password; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestHandler_Signup_CreatorError(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		err: errors.New("mock email reader error"),
	}
	mockHasher := &mockHasher{
		hash: "dummy-hashed-password",
		err:  nil,
	}
	mockCreator := &mockCreator{
		id:  "1234567890",
		err: errors.New("mock creator error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reqBody = struct {
		User     types.User
		Password string
	}{
		User: types.User{
			Name:  "Gabriel do Souza Seibel",
			Email: "gabriel.seibel@tuta.io",
		},
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
	New(mockHasher, nil, mockCreator, nil, mockByEmailReader, nil, nil, nil).Signup()(c)

	// assertions
	var resp struct {
		Error string
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Error, "email is taken"; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusUnprocessableEntity; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.User.Email; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := mockHasher.password, reqBody.Password; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := mockCreator.user.Name, reqBody.User.Name; got != want {
		t.Errorf("got created name %s, want %s", got, want)
	}
	if got, want := mockCreator.user.Email, reqBody.User.Email; got != want {
		t.Errorf("got created email %s, want %s", got, want)
	}
	if got, want := mockCreator.password, mockHasher.hash; got != want {
		t.Errorf("got created password %s, want %s", got, want)
	}
	if got, want := mockCreator.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
}

type mockVerifier struct {
	// receive
	hash     string
	password string

	// return
	err error
}

func (m *mockVerifier) CompareHashAndPassword(hashedPassword, password string) error {
	m.hash = hashedPassword
	m.password = password
	return m.err
}

func TestHandler_Login_OK(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		user: types.UserWithHashedPassword{
			User: types.User{
				ID:    "dummyID",
				Name:  "dummyName",
				Email: "dummyEmail",
			},
			HashedPassword: "dummyHash",
		},
		err: nil,
	}
	mockVerifier := &mockVerifier{
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
	New(nil, mockVerifier, nil, nil, mockByEmailReader, nil, nil, jwtSecret).Login()(c)

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
	if got, want := tokenName, mockByEmailReader.user.Name; got != want {
		t.Errorf("got token claim \"name\": %s, want %s", got, want)
	}
	tokenEmail := claims["email"].(string)
	if got, want := tokenEmail, mockByEmailReader.user.Email; got != want {
		t.Errorf("got token claim \"email\": %s, want %s", got, want)
	}
	tokenSub := claims["sub"].(string)
	if got, want := tokenSub, mockByEmailReader.user.ID; got != want {
		t.Errorf("got token claim \"sub\": %s, want %s", got, want)
	}
	tokenExp := claims["exp"].(float64)
	wantExp := time.Now().Add(time.Hour * 24 * 7)
	gotExp := time.Unix(int64(tokenExp), 0)
	if wantExp.Sub(gotExp).Abs() > time.Second*1 {
		t.Errorf("got token claim \"exp\": %s, want 1s from %s", gotExp, wantExp)
	}

	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.Email; got != want {
		t.Errorf("got loginer email = %s, want %s", got, want)
	}
	if got, want := mockVerifier.password, reqBody.Password; got != want {
		t.Errorf("got loginer password = %s, want %s", got, want)
	}
	if got, want := mockVerifier.hash, mockByEmailReader.user.HashedPassword; got != want {
		t.Errorf("got loginer password = %s, want %s", got, want)
	}
}

func TestHandler_Login_MissingEmail(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		user: types.UserWithHashedPassword{
			User: types.User{
				ID:    "dummyID",
				Name:  "dummyName",
				Email: "dummyEmail",
			},
			HashedPassword: "dummyHash",
		},
		err: nil,
	}
	mockVerifier := &mockVerifier{
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
	New(nil, mockVerifier, nil, nil, mockByEmailReader, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body error: %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Login_MissingPassword(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		user: types.UserWithHashedPassword{
			User: types.User{
				ID:    "dummyID",
				Name:  "dummyName",
				Email: "dummyEmail",
			},
			HashedPassword: "dummyHash",
		},
		err: nil,
	}
	mockVerifier := &mockVerifier{
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
	New(nil, mockVerifier, nil, nil, mockByEmailReader, nil, nil, jwtSecret).Login()(c)

	// assertions
	var resp struct {
		Err string `json:"error"`
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.Err, "missing user data"; got != want {
		t.Errorf("got response body error: %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func TestHandler_Login_ReaderError(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		user: types.UserWithHashedPassword{
			User: types.User{
				ID:    "dummyID",
				Name:  "dummyName",
				Email: "dummyEmail",
			},
			HashedPassword: "dummyHash",
		},
		err: errors.New("mock reader error"),
	}
	mockVerifier := &mockVerifier{
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
	New(nil, mockVerifier, nil, nil, mockByEmailReader, nil, nil, jwtSecret).Login()(c)

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
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.Email; got != want {
		t.Errorf("got loginer email = %s, want %s", got, want)
	}
}

func TestHandler_Login_VerifierError(t *testing.T) {
	// prepare test setup
	mockByEmailReader := &mockByEmailReader{
		user: types.UserWithHashedPassword{
			User: types.User{
				ID:    "dummyID",
				Name:  "dummyName",
				Email: "dummyEmail",
			},
			HashedPassword: "dummyHash",
		},
		err: nil,
	}
	mockVerifier := &mockVerifier{
		err: errors.New("mock verifier error"),
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
	New(nil, mockVerifier, nil, nil, mockByEmailReader, nil, nil, jwtSecret).Login()(c)

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
	if got, want := mockByEmailReader.ctx, c; got != want {
		t.Errorf("got passed context %v, want %v", got, want)
	}
	if got, want := mockByEmailReader.email, reqBody.Email; got != want {
		t.Errorf("got loginer email = %s, want %s", got, want)
	}
	if got, want := mockVerifier.password, reqBody.Password; got != want {
		t.Errorf("got loginer password = %s, want %s", got, want)
	}
	if got, want := mockVerifier.hash, mockByEmailReader.user.HashedPassword; got != want {
		t.Errorf("got loginer password = %s, want %s", got, want)
	}
}

func TestHandler_GetIDFromToken(t *testing.T) {
	// prepare test setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, nil, nil, nil, nil, nil, nil).GetIDFromToken()(c)

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

type mockByIDReader struct {
	// receive
	id  string
	ctx context.Context

	// return
	user types.User
	err  error
}

func (mr *mockByIDReader) ReadByID(ctx context.Context, id string) (types.User, error) {
	mr.id = id
	mr.ctx = ctx
	return mr.user, mr.err
}

func TestHandler_GetUserFromID_OK(t *testing.T) {
	// prepare test setup
	mockReader := &mockByIDReader{
		user: types.User{
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
	New(nil, nil, nil, mockReader, nil, nil, nil, nil).GetUserFromID()(c)

	// assertions
	var resp struct {
		User types.User `json:"user"`
	}
	err := json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.User, mockReader.user; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockReader.id, dummyID; got != want {
		t.Errorf("mockByIDReader.Read() received id %s, want %s", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Errorf("mockByIDReader.Read() received id %v, want %v", got, want)
	}
}

func TestHandler_GetUserFromID_ReaderError(t *testing.T) {
	// prepare test setup
	mockReader := &mockByIDReader{
		user: types.User{
			ID:    "mockReaderID",
			Name:  "mockReaderName",
			Email: "mockReaderEmail",
		},
		err: errors.New("mock byIDReader error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)

	// run code under test
	New(nil, nil, nil, mockReader, nil, nil, nil, nil).GetUserFromID()(c)

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
		t.Errorf("mockByIDReader.Read() received id %s, want %s", got, want)
	}
	if got, want := mockReader.ctx, c; got != want {
		t.Errorf("mockByIDReader.Read() received id %v, want %v", got, want)
	}
}

type mockUpdater struct {
	// receive
	user types.User
	ctx  context.Context

	// return
	err error
}

func (mu *mockUpdater) Update(ctx context.Context, user types.User) error {
	mu.user = user
	mu.ctx = ctx
	return mu.err
}

func TestHandler_UpdateUser_OK(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		user: types.User{
			ID:   "mockReaderID",
			Name: "mockReaderName",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := types.User{
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
	New(nil, nil, nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

	// assertions
	var resp struct {
		User types.User
	}
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	if err != nil {
		t.Errorf("got error %s decoding response, want nil", err)
	}
	if got, want := resp.User, mockUpdater.user; got != want {
		t.Errorf("got response body id %s, want %s", got, want)
	}
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
	if got, want := mockUpdater.user, reqBody; got != want {
		t.Errorf("mockUpdater.Update() received user %s, want %s", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Errorf("mockUpdater.Update() received id %v, want %v", got, want)
	}
}

func TestHandler_UpdateUser_MismatchedIDs(t *testing.T) {
	// prepare test setup
	mockUpdater := &mockUpdater{
		user: types.User{
			ID:   "mockReaderID",
			Name: "mockReaderName",
		},
		err: nil,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("AuthenticatedUserID", "1")
	reqBody := types.User{
		ID:   "2",
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
	New(nil, nil, nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

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
		user: types.User{
			ID:   "mockReaderID",
			Name: "mockReaderName",
		},
		err: errors.New("mock updater error"),
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	dummyID := "dummyID"
	c.Set("AuthenticatedUserID", dummyID)
	reqBody := types.User{
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
	New(nil, nil, nil, nil, nil, mockUpdater, nil, nil).UpdateUser()(c)

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
	if got, want := mockUpdater.user, reqBody; got != want {
		t.Errorf("mockUpdater.Update() received user %s, want %s", got, want)
	}
	if got, want := mockUpdater.ctx, c; got != want {
		t.Errorf("mockUpdater.Update() received id %v, want %v", got, want)
	}
}

type mockDeleter struct {
	// receive
	id  string
	ctx context.Context

	// return
	err error
}

func (md *mockDeleter) Delete(ctx context.Context, id string) error {
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
	New(nil, nil, nil, nil, nil, nil, mockDeleter, nil).DeleteUser()(c)

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
	New(nil, nil, nil, nil, nil, nil, mockDeleter, nil).DeleteUser()(c)

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
