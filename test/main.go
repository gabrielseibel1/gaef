package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	fmt.Println("\nvvvvvvvvvv GAEF E2E TESTS vvvvvvvvvv")

	////////////////////////
	// TEST USERS SERVICE //
	////////////////////////

	name := "A"
	email := "a@gmail.com"
	password := "test123"

	// create user
	userID, err := signUp(name, email, password)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n1. created user with id %v\n", userID)

	// login
	token, err := login(email, password)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n2. logged in with token %v\n", token)

	// read u
	u, err := readUser(token, userID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n3. read user as %v\n", u)

	// update user
	u.Name = "B"
	u.Email = "b@gmail.com"
	u, err = updateUser(token, u)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n4. updated user to %v\n", u)

	// delete user
	message, err := deleteUser(token, userID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n5. deleted user with message: %v\n", message)

	// validate token
	tokenID, err := readToken(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n6. validated token with id %v\n", tokenID)

	/////////////////////////
	// TEST GROUPS SERVICE //
	/////////////////////////

	// create group with three users
	user1ID, err := signUp("A", "a@gmail.com", "test123a")
	if err != nil {
		panic(err)
	}
	user2ID, err := signUp("B", "b@gmail.com", "test123b")
	if err != nil {
		panic(err)
	}
	user3ID, err := signUp("C", "c@gmail.com", "test123c")
	if err != nil {
		panic(err)
	}
	token, err = login("a@gmail.com", "test123a")
	if err != nil {
		panic(err)
	}
	g := group{
		Name:        "G",
		PictureURL:  "example.com",
		Description: "Gg",
		Members: []user{
			{
				ID:   user1ID,
				Name: "A",
			},
			{
				ID:   user2ID,
				Name: "B",
			},
			{
				ID:   user3ID,
				Name: "C",
			},
		},
		Leaders: []user{
			{
				ID:   user1ID,
				Name: "A",
			},
			{
				ID:   user3ID,
				Name: "C",
			},
		},
	}
	g, err = createGroup(token, g)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n7. created group %v\n", g)

	// create another group for the collections queries
	// to have multiple elements in the results
	g.Name = "H"
	g.Description = "Hh"
	g, err = createGroup(token, g)
	if err != nil {
		panic(err)
	}

	// participating groups
	groups, err := participatingGroups(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n8. got participating groups %v\n", groups)

	// leading groups
	groups, err = leadingGroups(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n9. got leading groups %v\n", groups)

	// read group
	g, err = readGroup(token, groups[0].ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n10. read group %v\n", g)

	// read leading group
	g, err = readLeadingGroup(token, groups[0].ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n11. read leading group %v\n", g)

	// update group
	g.Name = "I"
	g.Description = "Ii"
	g.Members, g.Leaders = g.Leaders, g.Members
	g, err = updateGroup(token, g)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n12. updated group %v\n", g)

	// delete group
	message, err = deleteGroup(token, groups[0].ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n13. deleted group with message : %s\n", message)
	_, err = deleteGroup(token, groups[1].ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n^^^^^^^^^^ GAEF E2E TESTS ^^^^^^^^^^")
}

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
		return "", fmt.Errorf("delete user request returned status code %d", resp.StatusCode)
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

type group struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	PictureURL  string `json:"pictureUrl,omitempty"`
	Description string `json:"description,omitempty"`
	Members     []user `json:"members,omitempty"`
	Leaders     []user `json:"leaders,omitempty"`
}

func createGroup(token string, g group) (group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return group{}, err
	}
	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/",
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return group{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return group{}, fmt.Errorf("create group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func participatingGroups(token string) ([]group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/participating",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("participating groups request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Groups []group `json:"groups"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func leadingGroups(token string) ([]group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/leading",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leading groups request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Groups []group `json:"groups"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func readGroup(token string, id string) (group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/" + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return group{}, fmt.Errorf("read group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func readLeadingGroup(token string, id string) (group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/leading/" + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return group{}, fmt.Errorf("read leading group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func updateGroup(token string, g group) (group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return group{}, err
	}
	req := http.Request{
		Method: http.MethodPut,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/" + g.ID,
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+token)
	c := http.Client{}
	resp, err := c.Do(&req)
	if err != nil {
		return group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return group{}, fmt.Errorf("update group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func deleteGroup(token, id string) (string, error) {
	req := http.Request{
		Method: http.MethodDelete,
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8081",
			Path:   "/api/v0/groups/" + id,
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
		return "", fmt.Errorf("delete group request returned status code %d", resp.StatusCode)
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
