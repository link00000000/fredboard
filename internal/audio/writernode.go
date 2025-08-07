package audio

import (
	"context"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	w io.Writer
}

func (node *WriterNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) (err error) {
	ctx, span := telemetry.Tracer.Start(ctx, "audio.WriterNode.Tick")
	defer span.End()

	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
	}

	if len(outs) != 0 {
		return newInvalidConnectionConfigErr(node, connectionType_Out, 0, 0, len(outs))
	}

	var n int64
	n, err = io.Copy(node.w, ins[0])
	telemetry.Logger.DebugContext(ctx, "WriterNode copied data from input to writer", "n", n, "error", err)

	return err
}

func NewWriterNode(w io.Writer) *WriterNode {
	return &WriterNode{w: w}
}
