package io

import (
	"context"
	"io"
)

func CopyContext(ctx context.Context, dst io.Writer, src io.Reader) (n int64, err error) {
	buf := make([]byte, 0x8000)

	for {
		select {
		case <-ctx.Done():
			return n, ctx.Err()
		default:
			nn, rerr := src.Read(buf)
			n = int64(nn)
			buf = buf[:n]

			if n > 0 {
				_, werr := dst.Write(buf)

				if werr != nil {
					return n, werr
				}
			}

			if rerr != nil {
				return n, rerr
			}
		}
	}
}
