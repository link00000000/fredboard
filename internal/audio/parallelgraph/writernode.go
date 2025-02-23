package parallelgraph

import (
	"context"
	"fmt"
	"io"

	internal_io "accidentallycoded.com/fredboard/v3/internal/io"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	logger     *logging.Logger
	w          io.Writer
	errs       chan error
	isFlushing bool
}

func (node *WriterNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(ins))
	}

	if len(outs) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(outs))
	}

	go func() {
		defer func() {
			close(node.errs)
			node.logger.Debug("closed Errors() channel")
		}()

		for {
			node.logger.Debug("start copying")
			n, err := internal_io.CopyContext(ctx, node.w, ins[0])
			node.logger.Debug("finished copying", "n", n, "error", err)

			if err == io.EOF {
				if node.isFlushing {
					_, _ = internal_io.CopyContext(ctx, node.w, ins[0])
					node.logger.Debug("flushed node", "node", node)
					break
				}

				continue
			}

			if err != nil {
				node.errs <- fmt.Errorf("error while copying to WriterNode writer: %w", err)
				return
			}
		}
	}()

	return nil
}

func (node *WriterNode) Errors() <-chan error {
	return node.errs
}

func (node *WriterNode) FlushAndStop() {
	node.isFlushing = true
}

func NewWriterNode(logger *logging.Logger, w io.Writer) *WriterNode {
	return &WriterNode{logger: logger, w: w, errs: make(chan error, 1)}
}
