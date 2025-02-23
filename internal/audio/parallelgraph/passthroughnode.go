package parallelgraph

import (
	"context"
	"fmt"
	"io"

	internal_io "accidentallycoded.com/fredboard/v3/internal/io"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*PassthroughNode)(nil)

type PassthroughNode struct {
	logger     *logging.Logger
	errs       chan error
	isFlushing bool
}

func (node *PassthroughNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(1, 1, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(1, 1, len(outs))
	}

	go func() {
		defer func() {
			close(node.errs)
			node.logger.Debug("closed Errors() channel")
		}()

		for {
			node.logger.Debug("start copying")
			n, err := internal_io.CopyContext(ctx, outs[0], ins[0])
			node.logger.Debug("finished copying", "n", n, "error", err)

			if err == io.EOF {
				if node.isFlushing {
					node.logger.Debug("flushed node", "node", node)
					_, _ = internal_io.CopyContext(ctx, outs[0], ins[0])
					break
				}

				continue
			}

			if err != nil {
				node.errs <- fmt.Errorf("error while copying passthrough node input to output: %w", err)
				return
			}
		}
	}()

	return nil
}

func (node *PassthroughNode) Errors() <-chan error {
	return node.errs
}

func (node *PassthroughNode) FlushAndStop() {
	node.isFlushing = true
}

func NewPassthroughNode(logger *logging.Logger) *PassthroughNode {
	return &PassthroughNode{logger: logger, errs: make(chan error, 1)}
}
