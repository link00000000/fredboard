package audio

import (
	"context"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	logger *logging.Logger

	w   io.Writer
	err error
}

func (node *WriterNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) {
	ctx, span := telemetry.Tracer.Start(ctx, "WriterNode.Tick")
	defer span.End()

	node.err = nil

	if len(ins) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
		return
	}

	if len(outs) != 0 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 0, 0, len(outs))
		return
	}

	var n int64
	n, node.err = io.Copy(node.w, ins[0])

	telemetry.Logger.DebugContext(ctx, "WriterNode copied data from input to writer", "n", n, "error", node.err)
}

func (node *WriterNode) Err() error {
	return node.err
}

func NewWriterNode(logger *logging.Logger, w io.Writer) *WriterNode {
	return &WriterNode{logger: logger, w: w, err: nil}
}
