package openai

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

type initSSEClietOption struct {
	APIEndpoint string
	Connection  *http.Client
	Method      string
}

func (c *Client) initSSEClient(opt *initSSEClietOption) {
	sseClient := sse.NewClient(APIEndpoint)
	sseClient.Connection = c.HTTPClient
	sseClient.Method = opt.Method

	c.SSEClient = sseClient
}
