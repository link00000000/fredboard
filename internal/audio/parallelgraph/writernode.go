package parallelgraph

import (
	"context"
	"fmt"
	"io"

	internal_io "accidentallycoded.com/fredboard/v3/internal/io"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	w    io.Writer
	errs chan error
}

func (node *WriterNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(ins))
	}

	if len(outs) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(outs))
	}

	go func() {
		defer close(node.errs)

		_, err := internal_io.CopyContext(ctx, node.w, ins[0])
		if err != nil {
			node.errs <- fmt.Errorf("error while copying to WriterNode writer: %w", err)
		}
	}()

	return nil
}

func (node *WriterNode) Errors() <-chan error {
	return node.errs
}

func NewWriterNode(w io.Writer) *WriterNode {
	return &WriterNode{w: w, errs: make(chan error, 1)}
}
