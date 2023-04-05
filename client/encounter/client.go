package encounter

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

func (c Client) CreateEncounter(ctx context.Context, token string, e types.Encounter) (string, error) {
	var respBody struct{ ID string }
	err := request(ctx, http.MethodPost, c.URL, e, token, &respBody)
	return respBody.ID, err
}

func (c Client) GetUserEncounters(ctx context.Context, token string) ([]types.Encounter, error) {
	var respBody struct{ Encounters []types.Encounter }
	err := request(ctx, http.MethodGet, c.URL, nil, token, &respBody)
	return respBody.Encounters, err
}

func (c Client) GetEncounter(ctx context.Context, token string, id string) (types.Encounter, error) {
	var respBody struct{ Encounter types.Encounter }
	err := request(ctx, http.MethodGet, c.URL+id, nil, token, &respBody)
	return respBody.Encounter, err
}

func (c Client) UpdateEncounter(ctx context.Context, token string, e types.Encounter) (types.Encounter, error) {
	var respBody struct{ Encounter types.Encounter }
	err := request(ctx, http.MethodPut, c.URL+e.ID, e, token, &respBody)
	return respBody.Encounter, err
}

func (c Client) DeleteEncounter(ctx context.Context, token string, id string) (string, error) {
	var respBody struct{ ID string }
	err := request(ctx, http.MethodDelete, c.URL+id, nil, token, &respBody)
	return respBody.ID, err
}

func (c Client) ConfirmEncounter(ctx context.Context, token string, id string) (string, error) {
	var respBody struct{ ID string }
	err := request(ctx, http.MethodPost, c.URL+id+"/confirmation", nil, token, &respBody)
	return respBody.ID, err
}

func (c Client) DeclineEncounter(ctx context.Context, token string, id string) (string, error) {
	var respBody struct{ ID string }
	err := request(ctx, http.MethodDelete, c.URL+id+"/confirmation", nil, token, &respBody)
	return respBody.ID, err
}

func request(ctx context.Context, method string, url string, bodyObj any, token string, respBody any) error {
	// build request body
	var body io.Reader
	if bodyObj == nil {
		body = nil
	} else {
		reqBodyBytes, err := json.Marshal(bodyObj)
		if err != nil {
			return err
		}
		body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
	}

	// build request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// parse response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request(%v, %v, %v) returned status code %d", method, url, body, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(respBody)
}
