package openai

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shuntaka9576/oax/sse"
)

var (
	ErrorOpenAIUnauthorized = errors.New("OpenAIUnauthorized")
)

type customTransport struct {
	http.RoundTripper
	openAISettings OpenAISettings
}

type OpenAISettings struct {
	BearerToken    string
	OrganizationID string
}

type Client struct {
	client         *sse.HTTPClient
	openAISettings OpenAISettings
}

type InitClientOptions struct {
	APIKey         string
	OrganizationID string
}

func InitClient(opt *InitClientOptions) *Client {
	client := &http.Client{
		Transport: &customTransport{
			RoundTripper: http.DefaultTransport,
			openAISettings: OpenAISettings{
				BearerToken:    opt.APIKey,
				OrganizationID: opt.OrganizationID,
			},
		},
	}

	httpClientWithSSE := sse.HTTPClient{
		Client: client,
	}

	return &Client{
		client: &httpClientWithSSE,
	}
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.openAISettings.BearerToken))
	req.Header.Set("Content-Type", "application/json")

	if t.openAISettings.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization:", t.openAISettings.OrganizationID)
	}

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("OpenAI API Request error status code %d: %w", resp.StatusCode, ErrorOpenAIUnauthorized)
		} else {
			return nil, fmt.Errorf("OpenAI API Request error status code %d", resp.StatusCode)
		}
	} else {
		return resp, err
	}
}
