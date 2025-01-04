package server

import (
	"errors"
	"fmt"
	"net/http"
)

type SSEConnection struct {
	writer    http.ResponseWriter
	connected bool
}

func NewSSEConnection(w http.ResponseWriter) *SSEConnection {
	return &SSEConnection{writer: w, connected: false}
}

var ErrStreamingUnsupported = errors.New("streaming unsupported")

func (writer *SSEConnection) EstablishConnection() error {
	w := writer.writer

	// Flush the headers to establish the connection
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported!", http.StatusInternalServerError)
		return ErrStreamingUnsupported
	}

	if !writer.connected {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher.Flush()

		writer.connected = true
	}

	return nil
}

// Implements [io.Writer]
func (writer *SSEConnection) Write(p []byte) (int, error) {
	w := writer.writer

	n, err := fmt.Fprint(w, fmt.Sprintf("data: %s\n\n", p))

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported!", http.StatusInternalServerError)
		return 0, ErrStreamingUnsupported
	}
	flusher.Flush()

	return n, err
}

type SSEBroadcaster struct {
	writers map[int]*SSEConnection
	nextId  int
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{writers: make(map[int]*SSEConnection), nextId: 0}
}

func (broadcaster SSEBroadcaster) AddResponse(conn *SSEConnection) int {
	id := broadcaster.nextId
	broadcaster.writers[id] = conn

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
