package audio

import (
	"context"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	r    io.Reader
	size int64
}

func (node *ReaderNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) (err error) {
	ctx, span := telemetry.Tracer.Start(ctx, "audio.ReaderNode.Tick")
	defer span.End()

	if len(ins) != 0 {
		return newInvalidConnectionConfigErr(node, connectionType_In, 0, 0, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(node, connectionType_Out, 1, 1, len(outs))
	}

	var n int64
	n, err = io.CopyN(outs[0], node.r, node.size)
	telemetry.Logger.DebugContext(ctx, "ReaderNode copied data from reader to output", "n", n, "error", err)

	return err
}

func NewReaderNode(r io.Reader, size int64) *ReaderNode {
	return &ReaderNode{r: r, size: size}
}
