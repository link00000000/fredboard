package audio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"

	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

var _ Node = (*MixerNode)(nil)

type MixerNode struct {
}

func (node *MixerNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) (err error) {
	ctx, span := telemetry.Tracer.Start(ctx, "audio.MixerNode.Tick")
	defer span.End()

	if len(ins) <= 0 {
		return newInvalidConnectionConfigErr(node, connectionType_In, 0, connection_Unbounded, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(node, connectionType_Out, 1, 1, len(outs))
	}

	errs := make([]error, 0)
	mixedStream := make([]int16, 0) // TODO: cache

	for inIdx, in := range ins {
		bytes, err := io.ReadAll(in)
		telemetry.Logger.DebugContext(ctx, "MixerNode copied data from input to internal buffer", "input", inIdx, "n", len(bytes), "error", err)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read from input %d: %w", inIdx, err))
			continue
		}

		stream := codecs.BytesToS16LE(bytes)

		if len(stream) > len(mixedStream) {
			telemetry.Logger.DebugContext(ctx, "mixedStream buffer is too small. growing buffer", "oldcap", "newcap")

			mixedStream = slices.Grow(mixedStream, len(stream)-len(mixedStream))
			mixedStream = mixedStream[:cap(mixedStream)]
		}

		for sampleIdx, sample := range stream {
			i32 := int32(mixedStream[sampleIdx]) + int32(sample)
			switch {
			case i32 < math.MinInt16: // underflow so set the sample to the min value
				mixedStream[sampleIdx] = math.MinInt16
			case i32 > math.MaxInt16: // overflow so set the sample to the max value
				mixedStream[sampleIdx] = math.MaxInt16
			default:
				mixedStream[sampleIdx] = int16(i32)
			}
		}
	}

	n, err := outs[0].Write(codecs.S16LEToBytes(mixedStream))
	telemetry.Logger.DebugContext(ctx, "MixerNode copied data from internal buffer to output", "n", n, "error", err)

	if err != nil {
		errs = append(errs, fmt.Errorf("failed to copy data from internal buffer to output: %w", err))
	}

	return errors.Join(errs...)
}

func NewMixerNode() *MixerNode {
	return &MixerNode{}
}
