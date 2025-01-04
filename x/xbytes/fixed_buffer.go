package xbytes

import (
	"errors"
	"io"
)

var (
	ErrBufferFull = errors.New("buffer is full")
)

var (
	_ io.Reader = (*FixedBuffer)(nil)
	_ io.Writer = (*FixedBuffer)(nil)
)

type FixedBuffer struct {
	buf    []byte
	offset int
}

func (buffer *FixedBuffer) Read(p []byte) (int, error) {
	if buffer.offset == 0 {
		return 0, io.EOF
	}

	n := copy(p, buffer.buf)
	buffer.buf = buffer.buf[n:]
	buffer.offset -= n

	return n, nil
}

func (buffer *FixedBuffer) Write(p []byte) (int, error) {
	remainingCapacity := len(buffer.buf) - buffer.offset
	if remainingCapacity == 0 {
		return 0, ErrBufferFull
	}

	n := copy(buffer.buf[buffer.offset:], p)
	buffer.offset += n

	return n, nil
}

func NewFixedBuffer(size int) *FixedBuffer {
	return &FixedBuffer{buf: make([]byte, size), offset: 0}
}
