package encounterProposal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gabrielseibel1/gaef/types"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	URL string
}

func (c Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"health", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health request returned status code %d", resp.StatusCode)
	}
	return nil
}

func (c Client) CreateEP(ctx context.Context, token string, ep types.EncounterProposal) (string, error) {
	reqBodyBytes, err := json.Marshal(ep)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create EP request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ ID string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.ID, err
}

func (c Client) Mine(ctx context.Context, token string) ([]types.EncounterProposal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"mine", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mine eps request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ EncounterProposals []types.EncounterProposal }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.EncounterProposals, err
}

func (c Client) Page(ctx context.Context, token string, page int) ([]types.EncounterProposal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"page/"+strconv.Itoa(page), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("page eps request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ EncounterProposals []types.EncounterProposal }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.EncounterProposals, err
}

func (c Client) ReadEP(ctx context.Context, token string, id string) (types.EncounterProposal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+id, nil)
	if err != nil {
		return types.EncounterProposal{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.EncounterProposal{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.EncounterProposal{}, fmt.Errorf("read ep request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ EncounterProposal types.EncounterProposal }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.EncounterProposal, err
}

func (c Client) UpdateEP(ctx context.Context, token string, ep types.EncounterProposal) (types.EncounterProposal, error) {
	reqBodyBytes, err := json.Marshal(ep)
	if err != nil {
		return types.EncounterProposal{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.URL+ep.ID, io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return types.EncounterProposal{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.EncounterProposal{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return types.EncounterProposal{}, fmt.Errorf("update EP request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ EncounterProposal types.EncounterProposal }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.EncounterProposal, err
}

func (c Client) DeleteEP(ctx context.Context, token string, id string) (string, error) {
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
		return "", fmt.Errorf("delete EP request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Message string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Message, err
}

func (c Client) ApplyToEP(ctx context.Context, token string, id string, app types.Application) (string, error) {
	reqBodyBytes, err := json.Marshal(app)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+id+"/applications", io.NopCloser(bytes.NewBuffer(reqBodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("apply to EP request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Message string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Message, err
}

func (c Client) DeleteApplication(ctx context.Context, token string, epID string, appID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.URL+epID+"/applications/"+appID, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("delete application apply to EP request returned status code %d", resp.StatusCode)
	}

	var respBody struct{ Message string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	return respBody.Message, err
}
