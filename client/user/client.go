package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gabrielseibel1/gaef/client/domain"
)

type Client struct {
	Host       string
	BasePath   string
	Token      string
	HTTPClient http.Client
}

func (c Client) SignUp(name, email, password string) (string, error) {
	reqBodyStruct := domain.User{
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
			Host:   c.Host,
			Path:   c.BasePath,
		},
		Body: io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
	}
	resp, err := c.HTTPClient.Do(&req)
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

func (c Client) Login(email, password string) (string, error) {
	reqBodyStruct := domain.User{
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
			Host:   c.Host,
			Path:   c.BasePath + "session",
		},
		Body: io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
	}
	resp, err := c.HTTPClient.Do(&req)
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

func (c Client) ReadUser(id string) (domain.User, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return domain.User{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return domain.User{}, fmt.Errorf("read user request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		User domain.User `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func (c Client) UpdateUser(u domain.User) (domain.User, error) {
	reqBodyBytes, err := json.Marshal(u)
	if err != nil {
		return u, err
	}
	req := http.Request{
		Method: http.MethodPut,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + u.ID,
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return u, err
	}
	if resp.StatusCode != http.StatusOK {
		return u, fmt.Errorf("update user request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		User domain.User `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func (c Client) DeleteUser(id string) (string, error) {
	req := http.Request{
		Method: http.MethodDelete,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
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

func (c Client) ReadToken() (string, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + "token-validation",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
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
