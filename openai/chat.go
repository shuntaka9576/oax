package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/shuntaka9576/oax/sse"
)

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Delta        Message     `json:"delta"`
	Index        int         `json:"index"`
	FinishReason interface{} `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatCreateCompletionOption struct {
	Model    string
	Messages []Message
}

type CreateCompletionStream struct {
}

type CreateCompletionStreamEvent struct {
}

func (c *Client) ChatCreateCompletionSubscribeWithContext(ctx context.Context, opt *ChatCreateCompletionOption, handler func(msg *ChatCompletionResponse, err error)) error {
	body := requestBody{
		Messages: opt.Messages,
		Model:    opt.Model,
		Stream:   true,
	}

	reqBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	c.client.SubscribeWithContext(ctx, APIBaseEndpoint+"/v1/chat/completions", "POST", bytes.NewBuffer(reqBytes), func(msg *sse.Event, err error) {
		if err != nil {

			handler(nil, err)
			return
		}

		if msg != nil {
			if string(msg.Data) == "[DONE]" {
				handler(nil, io.EOF)

				return
			} else {
				var jsonData ChatCompletionResponse
				err = json.Unmarshal([]byte(msg.Data), &jsonData)
				if err != nil {
					handler(nil, err)
				}

				handler(&jsonData, err)
			}
		}
	})

	return nil
}
