package parallelgraph

import (
	"context"
	"fmt"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/events"
	internal_io "accidentallycoded.com/fredboard/v3/internal/io"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	logger     *logging.Logger
	r          io.Reader
	errs       chan error
	isFlushing bool

	OnEOF *events.EventEmitter[struct{}]
}

func (node *ReaderNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(outs))
	}

	go func() {
		defer func() {
			close(node.errs)
			node.logger.Debug("closed Errors() channel")
		}()

		for {
			node.logger.Debug("start copying")
			n, err := internal_io.CopyContext(ctx, outs[0], node.r)
			node.logger.Debug("finished copying", "n", n, "error", err)

			if err == io.EOF {
				if node.isFlushing {
					node.logger.Debug("flushed node", "node", node)
					_, _ = internal_io.CopyContext(ctx, outs[0], node.r)
					break
				}

				// TODO: Replace with something else that signifies that the node is flushed and done processing
				node.OnEOF.Broadcast(struct{}{})
				continue
			}

			if err != nil {
				node.errs <- fmt.Errorf("error while copying from ReaderNode reader: %w", err)
				return
			}
		}
	}()

	return nil
}

func (node *ReaderNode) Errors() <-chan error {
	return node.errs
}

func (node *ReaderNode) FlushAndStop() {
	node.isFlushing = true
}

func NewReaderNode(logger *logging.Logger, r io.Reader) *ReaderNode {
	return &ReaderNode{logger: logger, r: r, errs: make(chan error, 1), OnEOF: events.NewEventEmitter[struct{}]()}
}
