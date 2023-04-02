package sse

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

type EventStreamReader struct {
	scanner *bufio.Scanner
}

func NewEventStreamReader(reader io.Reader) EventStreamReader {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 4096)

	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i, nlen := containsDoubleNewline(data); i >= 0 {
			return i + nlen, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	scanner.Split(split)

	return EventStreamReader{
		scanner: scanner,
	}
}

func (e *EventStreamReader) ReadEvent() ([]byte, error) {

	if e.scanner.Scan() {
		event := e.scanner.Bytes()
		return event, nil
	}
	if err := e.scanner.Err(); err != nil {
		if err == context.Canceled {
			return nil, io.EOF
		}
		return nil, err
	}

	return nil, io.EOF
}

func containsDoubleNewline(data []byte) (int, int) {
	lflfPos := bytes.Index(data, []byte("\n\n"))

	return lflfPos, 2
}
