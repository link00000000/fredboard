package parallelgraph

import (
	"context"
	"fmt"
	"io"

	internal_io "accidentallycoded.com/fredboard/v3/internal/io"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	r    io.Reader
	errs chan error
}

func (node *ReaderNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(outs))
	}

	go func() {
		defer close(node.errs)

		_, err := internal_io.CopyContext(ctx, outs[0], node.r)
		if err != nil {
			node.errs <- fmt.Errorf("error while copying from ReaderNode reader: %w", err)
		}
	}()

	return nil
}

func (node *ReaderNode) Errors() <-chan error {
	return node.errs
}

func NewReaderNode(r io.Reader) *ReaderNode {
	return &ReaderNode{r: r, errs: make(chan error, 1)}
}
