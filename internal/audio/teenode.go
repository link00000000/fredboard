package audio

import (
	"context"
	"errors"
	"fmt"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

var _ Node = (*TeeNode)(nil)

type TeeNode struct {
}

func (node *TeeNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) (err error) {
	ctx, span := telemetry.Tracer.Start(ctx, "audio.TeeNode.Tick")
	defer span.End()

	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
	}

	if len(outs) <= 0 {
		return newInvalidConnectionConfigErr(node, connectionType_Out, 0, connection_Unbounded, len(outs))
	}

	bytes, err := io.ReadAll(ins[0])
	telemetry.Logger.DebugContext(ctx, "TeeNode copied data from input to internal buffer", "n", len(bytes), "error", err)

	if err != nil {
		return fmt.Errorf("failed to read from input: %w", err)
	}

	errs := make([]error, 0)

	for outIdx, out := range outs {
		n, err := out.Write(bytes)
		telemetry.Logger.DebugContext(ctx, "TeeNode copied data from input to internal buffer", "n", n, "error", err)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to write data to output %d: %w", outIdx, err))
			continue
		}
	}

	return errors.Join(errs...)
}

func NewTeeNode() *TeeNode {
	return &TeeNode{}
}
