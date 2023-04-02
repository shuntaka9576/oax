package sse

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

var (
	headerData = []byte("data:")
)

type HTTPClient struct {
	Client *http.Client
}

type Event struct {
	Data []byte
}

func (c *HTTPClient) sseRequest(ctx context.Context, url string, method string, request io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, request)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return c.Client.Do(req)
}

func (c *HTTPClient) SubscribeWithContext(ctx context.Context, url string, method string, request io.Reader, handler func(msg *Event, err error)) error {
	res, err := c.sseRequest(ctx, url, method, request)
	if err != nil {
		handler(nil, err)

		return err
	}
	defer res.Body.Close()

	eventStreamReader := NewEventStreamReader(res.Body)

	eventCh := make(chan *Event)
	errCh := make(chan error)

	go func() {
		for {
			eventBytes, err := eventStreamReader.ReadEvent()
			if err != nil {
				if err == io.EOF {
					errCh <- nil
					return
				}

				errCh <- err
				return
			}

			var e Event
			for _, line := range bytes.FieldsFunc(eventBytes, func(r rune) bool { return r == '\n' || r == '\r' }) {

				switch {
				case bytes.HasPrefix(line, headerData):
					e.Data = append(e.Data[:], append(trimHeader(len(headerData), line), byte('\n'))...)
				}
			}

			e.Data = bytes.TrimSuffix(e.Data, []byte("\n"))
			eventCh <- &e
		}
	}()

	for {
		select {
		case err = <-errCh:
			handler(nil, err)

			return err
		case msg := <-eventCh:
			handler(msg, nil)
		case <-ctx.Done():
			return nil
		}
	}
}

func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}

	data = data[size:]
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	if len(data) > 0 && data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}

	return data
}
