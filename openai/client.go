package openai

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/r3labs/sse/v2"
)

type Client struct {
	HTTPClient *http.Client
	SSEClient  *sse.Client
}

type InitClientOptions struct {
	APIKey         string
	OrganizationID string
	ErrCh          chan error
}

type customTransport struct {
	http.RoundTripper
	APIKey         string
	OrganizationID string
	ErrCh          chan error
}

func InitClient(opt *InitClientOptions) *Client {
	client := &http.Client{
		Transport: &customTransport{
			RoundTripper:   http.DefaultTransport,
			APIKey:         opt.APIKey,
			OrganizationID: opt.OrganizationID,
			ErrCh:          opt.ErrCh,
		},
	}

	return &Client{
		HTTPClient: client,
	}
}

var (
	ErrorOpenAIUnauthorized = errors.New("OpenAIUnauthorized")
)

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.APIKey))
	req.Header.Set("Content-Type", "application/json")

	if t.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization:", t.OrganizationID)
	}

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		t.ErrCh <- err

		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 401 {
			t.ErrCh <- fmt.Errorf("OpenAI API Request error status code %d: %w", resp.StatusCode, ErrorOpenAIUnauthorized)
		} else {
			t.ErrCh <- fmt.Errorf("OpenAI API Request error status code %d", resp.StatusCode)
		}

		return nil, err
	} else {
		return resp, err
	}
}
