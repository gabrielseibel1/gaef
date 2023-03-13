package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type user struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func signUp(name, email, password string) (string, error) {
	reqBodyStruct := user{
		Name:     name,
		Email:    email,
		Password: password,
	}
	reqBodyBytes, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return "", err
	}
	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/",
		},
		Body: io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
	}
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("sign-up request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.ID, err
}

func login(email, password string) (string, error) {
	reqBodyStruct := user{
		Email:    email,
		Password: password,
	}
	reqBodyBytes, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return "", err
	}
	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/session",
		},
		Body: io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
	}
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Token, nil
}

func readUser(token, id string) (user, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/" + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return user{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return user{}, fmt.Errorf("read user request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		User user `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func updateUser(token string, u user) (user, error) {
	reqBodyBytes, err := json.Marshal(u)
	if err != nil {
		return u, err
	}
	req := http.Request{
		Method: http.MethodPut,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/" + u.ID,
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return u, err
	}
	if resp.StatusCode != http.StatusOK {
		return u, fmt.Errorf("update user request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		User user `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func deleteUser(token, id string) (string, error) {
	req := http.Request{
		Method: http.MethodDelete,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/" + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "user{}", fmt.Errorf("delete user request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Message string `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Message, nil
}

func readToken(token string) (string, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/api/v0/users/token-validation",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("read token request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.ID, err
}

func main() {
	fmt.Println("\nvvvvvvvvvv GAEF E2E TESTS vvvvvvvvvv")

	name := "Gabriel de Souza Seibel"
	email := "gabriel.seibel@tuta.io"
	password := "test123"

	userID, err := signUp(name, email, password)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n1. created user with id %v\n", userID)

	token, err := login(email, password)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n2. logged in with token %v\n", token)

	user, err := readUser(token, userID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n3. read user as %v\n", user)

	user.Name = "Gabriel Seibel de Souza"
	user.Email = "gabrielseibel1@gmail.com"
	user, err = updateUser(token, user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n4. updated user to %v\n", user)

	message, err := deleteUser(token, userID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n5. deleted user with message: %v\n", message)

	tokenID, err := readToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n6. validated token with id %v\n", tokenID)

	fmt.Println("\n^^^^^^^^^^ GAEF E2E TESTS ^^^^^^^^^^")
}
