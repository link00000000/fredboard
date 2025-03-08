package ioext

import (
	"io"
)

type channelWriter struct {
	c chan<- []byte
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	w.c <- p
	return len(p), err
}

func (w *channelWriter) Close() error {
	close(w.c)
	return nil
}

func NewChannelWriter(c chan<- []byte) io.WriteCloser {
	return &channelWriter{c: c}
}
