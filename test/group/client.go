package group

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gabrielseibel1/gaef-test/domain"
)

type Client struct {
	Host       string
	BasePath   string
	Token      string
	HTTPClient http.Client
}

func (c Client) CreateGroup(g domain.Group) (domain.Group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return domain.Group{}, err
	}
	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath,
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return domain.Group{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return domain.Group{}, fmt.Errorf("create group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group domain.Group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) ParticipatingGroups() ([]domain.Group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + "participating",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("participating groups request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Groups []domain.Group `json:"groups"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func (c Client) LeadingGroups() ([]domain.Group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + "leading",
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leading groups request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Groups []domain.Group `json:"groups"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Groups, err
}

func (c Client) ReadGroup(id string) (domain.Group, error) {
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
		return domain.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Group{}, fmt.Errorf("read group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group domain.Group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) ReadLeadingGroup(id string) (domain.Group, error) {
	req := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + "leading/" + id,
		},
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return domain.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Group{}, fmt.Errorf("read leading group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group domain.Group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) UpdateGroup(g domain.Group) (domain.Group, error) {
	reqBodyBytes, err := json.Marshal(g)
	if err != nil {
		return domain.Group{}, err
	}
	req := http.Request{
		Method: http.MethodPut,
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Host,
			Path:   c.BasePath + g.ID,
		},
		Body:   io.NopCloser(bytes.NewBuffer(reqBodyBytes)),
		Header: make(http.Header),
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(&req)
	if err != nil {
		return domain.Group{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Group{}, fmt.Errorf("update group request returned status code %d", resp.StatusCode)
	}
	var respBody struct {
		Group domain.Group `json:"group"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Group, err
}

func (c Client) DeleteGroup(id string) (string, error) {
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
