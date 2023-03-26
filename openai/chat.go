package openai

import (
	"bytes"
	"encoding/json"

	"github.com/r3labs/sse/v2"
)

type Choice struct {
	Delta        Message     `json:"delta"`
	Index        int         `json:"index"`
	FinishReason interface{} `json:"finish_reason"`
}

type JSONData struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

const APIEndpoint = "https://api.openai.com/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestBody struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

type ChatCreateCompletionOption struct {
	Model     string
	Messages  []Message
	MessageCh chan Message
	DoneCh    chan struct{}
	ErrCh     chan error
}

func (c *Client) ChatCreateCompletion(opt *ChatCreateCompletionOption) error {
	c.initSSEClient(&initSSEClietOption{
		APIEndpoint: APIBaseEndpoint + "/v1/chat/completions",
		Connection:  c.HTTPClient,
		Method:      "POST",
	})

	body := requestBody{
		Messages: opt.Messages,
		Model:    opt.Model,
		Stream:   true,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	c.SSEClient.Body = bytes.NewBuffer([]byte(jsonData))

	c.SSEClient.SubscribeRaw(func(msg *sse.Event) {
		var jsonData JSONData

		if string(msg.Data) == "[DONE]" {
			opt.DoneCh <- struct{}{}
		} else {
			err := json.Unmarshal([]byte(msg.Data), &jsonData)
			if err != nil {
				opt.ErrCh <- err
			}

			opt.MessageCh <- jsonData.Choices[0].Delta
		}
	})

	return nil
}
