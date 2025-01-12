package graph

import "bytes"

type AudioGraphConnectionBuffer struct {
	buf []byte
}

type AudioGraphConnection struct {
	from AudioGraphNode
	to   AudioGraphNode
	buf  bytes.Buffer
}

// Implements [io.Reader]
func (conn *AudioGraphConnection) Read(p []byte) (int, error) {
	return conn.buf.Read(p)
}

// Implements [io.Writer]
func (conn *AudioGraphConnection) Write(p []byte) (int, error) {
	return conn.buf.Write(p)
}

func NewAudioGraphConnection(from, to AudioGraphNode) *AudioGraphConnection {
	return &AudioGraphConnection{from: from, to: to}
}
