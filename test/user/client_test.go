package user

import (
	"net/http"
	"testing"
)

func TestClient_CRUD_Localhost8080(t *testing.T) {
	usersClient := Client{
		Host:       "localhost:8080",
		BasePath:   "/api/v0/users/",
		HTTPClient: http.Client{},
	}

	name := "A"
	email := "a@gmail.com"
	password := "test123"

	// create user
	userID, err := usersClient.SignUp(name, email, password)
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}

	// login
	token, err := usersClient.Login(email, password)
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	usersClient.Token = token

	// read user
	u, err := usersClient.ReadUser(userID)
	if err != nil {
		t.Fatalf("usersClient.ReadUser = err: %s", err.Error())
	}

	// update user
	u.Name = "B"
	u.Email = "b@gmail.com"
	u, err = usersClient.UpdateUser(u)
	if err != nil {
		t.Fatalf("usersClient.UpdateUser = err: %s", err.Error())
	}

	// delete user
	_, err = usersClient.DeleteUser(userID)
	if err != nil {
		t.Fatalf("usersClient.DeleteUser = err: %s", err.Error())
	}

	// validate token
	_, err = usersClient.ReadToken()
	if err != nil {
		t.Fatalf("usersClient.ReadToken = err: %s", err.Error())
	}
}
