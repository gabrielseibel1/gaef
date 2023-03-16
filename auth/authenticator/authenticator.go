package authenticator

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Authenticator struct {
	userServiceURL string
}

func New(userServiceURL string) Authenticator {
	return Authenticator{
		userServiceURL: userServiceURL,
	}
}

func (a Authenticator) GetAuthenticatedUserID(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.userServiceURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("unauthorized")
	}
	var payload userServiceResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return "", errors.New("unauthorized")
	}
	if payload.ErrMessage != "" {
		return "", errors.New("unauthorized")
	}

	return payload.ID, nil
}

type userServiceResponse struct {
	ID         string `json:"id,omitempty"`
	ErrMessage string `json:"error,omitempty"`
}
