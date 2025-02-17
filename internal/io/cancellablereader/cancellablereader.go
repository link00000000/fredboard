package cancellablereader

import (
	"context"
	"fmt"
	"io"
)

var _ io.ReadCloser = (*CancellableReader)(nil)

type CancellableReader struct {
	ctx  context.Context
	data chan []byte
	err  error
	r    io.Reader
}

func (cr *CancellableReader) startRead() {
	buf := make([]byte, 1024)
	for {
		n, err := cr.r.Read(buf)
		buf = buf[:n]

		if n > 0 {
			tmp := make([]byte, n)
			copy(tmp, buf)
			cr.data <- tmp
		}

		if err != nil {
			cr.err = err
			close(cr.data)
			return
		}
	}
}

func (cr *CancellableReader) Read(p []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, fmt.Errorf("reader cancelled: %w", cr.ctx.Err())
	case d, ok := <-cr.data:
		if !ok {
			return 0, cr.err
		}
		n := copy(p, d)
		return n, nil
	}
}

func (cr *CancellableReader) Close() (err error) {
	if closer, ok := cr.r.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func New(ctx context.Context, r io.Reader) (cr *CancellableReader) {
	cr = &CancellableReader{
		ctx:  ctx,
		data: make(chan []byte),
		r:    r,
	}

	go cr.startRead()
	return cr
}
