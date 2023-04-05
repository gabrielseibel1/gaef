package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gabrielseibel1/gaef/types"
	"io"
	"net/http"
)

type Client struct {
	URL string
}

func (c Client) SignUp(ctx context.Context, user types.User, password string) (string, error) {
	var body = struct {
		User     types.User
		Password string
	}{user, password}
	reqBodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("sign-up request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ ID string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.ID, err
}

func (c Client) Login(ctx context.Context, email, password string) (string, error) {
	var body = struct {
		Email    string
		Password string
	}{
		email, password,
	}
	reqBodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"session", io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Token string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Token, nil
}

func (c Client) ReadUser(ctx context.Context, token, id string) (types.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+id, nil)
	if err != nil {
		return types.User{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.User{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.User{}, fmt.Errorf("read user request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ User types.User }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func (c Client) UpdateUser(ctx context.Context, token string, u types.User) (types.User, error) {
	reqBodyBytes, err := json.Marshal(u)
	if err != nil {
		return u, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.URL+u.ID, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return types.User{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return u, err
	}
	if resp.StatusCode != http.StatusOK {
		return u, fmt.Errorf("update user request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ User types.User }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.User, err
}

func (c Client) DeleteUser(ctx context.Context, token, id string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.URL+id, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("delete user request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Message string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Message, nil
}

func (c Client) ReadToken(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"token-validation", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("read token request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ ID string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.ID, err
}
