package user_test

import (
	"context"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/types"
	"testing"
)

func TestClient_CRUD_Localhost8080(t *testing.T) {
	ctx := context.TODO()

	usersClient := user.Client{URL: "http://localhost:8080/api/v0/users/"}

	name := "A"
	email := "a@gmail.com"
	password := "test123"

	// create user
	userID, err := usersClient.SignUp(ctx, types.User{Name: name, Email: email}, password)
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}

	// login
	token, err := usersClient.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}

	// read user
	u, err := usersClient.ReadUser(ctx, token, userID)
	if err != nil {
		t.Fatalf("usersClient.ReadUser = err: %s", err.Error())
	}

	// update user
	u.Name = "B"
	u.Email = "b@gmail.com"
	_, err = usersClient.UpdateUser(ctx, token, u)
	if err != nil {
		t.Fatalf("usersClient.UpdateUser = err: %s", err.Error())
	}

	// delete user
	_, err = usersClient.DeleteUser(ctx, token, userID)
	if err != nil {
		t.Fatalf("usersClient.DeleteUser = err: %s", err.Error())
	}

	// validate token
	_, err = usersClient.ReadToken(ctx, token)
	if err != nil {
		t.Fatalf("usersClient.ReadToken = err: %s", err.Error())
	}
}
