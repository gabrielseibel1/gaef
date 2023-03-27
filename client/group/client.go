package group

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

func (c Client) CreateGroup(ctx context.Context, token string, g types.Group) (types.Group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return types.Group{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return types.Group{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Group{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return types.Group{}, fmt.Errorf("create group request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Group types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) ParticipatingGroups(ctx context.Context, token string) ([]types.Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"participating", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("participating groups request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Groups []types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func (c Client) LeadingGroups(ctx context.Context, token string) ([]types.Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"leading", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leading groups request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Groups []types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func (c Client) ReadGroup(ctx context.Context, token, id string) (types.Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+id, nil)
	if err != nil {
		return types.Group{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.Group{}, fmt.Errorf("read group request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Group types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) ReadLeadingGroup(ctx context.Context, token, id string) (types.Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"leading/"+id, nil)
	if err != nil {
		return types.Group{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.Group{}, fmt.Errorf("read leading group request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Group types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) UpdateGroup(ctx context.Context, token string, g types.Group) (types.Group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return types.Group{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.URL+g.ID, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return types.Group{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.Group{}, fmt.Errorf("update group request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Group types.Group }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) DeleteGroup(ctx context.Context, token, id string) (string, error) {
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
		return "", fmt.Errorf("delete group request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Message string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Message, nil
}

func (c Client) IsGroupLeader(ctx context.Context, token, groupID string) (bool, error) {
	_, err := c.ReadLeadingGroup(ctx, token, groupID)
	return err == nil, err
}
