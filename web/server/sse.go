package server

import (
	"errors"
	"fmt"
	"net/http"
)

type SSEResponseWriter struct {
	res       http.ResponseWriter
	connected bool
}

func NewSSEResponseWriter(res http.ResponseWriter) SSEResponseWriter {
	return SSEResponseWriter{res: res, connected: false}
}

var ErrStreamingUnsupported = errors.New("streaming unsupported")

// Implements [io.Writer]
func (writer *SSEResponseWriter) Write(p []byte) (int, error) {
	res := writer.res

	// Flush the headers to establish the connection
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "streaming unsupported!", http.StatusInternalServerError)
		return 0, ErrStreamingUnsupported
	}

	if !writer.connected {
		res.Header().Set("Content-Type", "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")

		flusher.Flush()

		writer.connected = true
	}

	n, err := fmt.Fprint(res, fmt.Sprintf("data: %s\n\n", p))

	flusher.Flush()

	return n, err
}

type SSEBroadcaster struct {
	writers map[int]SSEResponseWriter
	nextId  int
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{writers: make(map[int]SSEResponseWriter), nextId: 0}
}

func (broadcaster SSEBroadcaster) AddResponse(res http.ResponseWriter) int {
	id := broadcaster.nextId
	broadcaster.writers[id] = NewSSEResponseWriter(res)

	broadcaster.nextId++

	return id
}

func (broadcaster SSEBroadcaster) RemoveResponse(id int) {
	delete(broadcaster.writers, id)
}

// Implements [io.Writer]
func (broadcaster SSEBroadcaster) Write(p []byte) (int, error) {
	errs := make([]error, 0)
	for _, writer := range broadcaster.writers {
		_, err := writer.Write(p)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return len(p), errors.Join(errs...)
	}

	return len(p), nil
}
