package bytes

import (
	"bytes"
	"sync"
)

type ThreadsafeBuffer struct {
	buffer bytes.Buffer
	m      sync.Mutex
}

func (tb *ThreadsafeBuffer) Write(p []byte) (n int, err error) {
	tb.m.Lock()
	defer tb.m.Unlock()

	return tb.buffer.Write(p)
}

func (tb *ThreadsafeBuffer) Read(p []byte) (n int, err error) {
	tb.m.Lock()
	defer tb.m.Unlock()

	return tb.buffer.Read(p)
}

func (tb *ThreadsafeBuffer) Bytes() []byte {
	return tb.Bytes()
}
