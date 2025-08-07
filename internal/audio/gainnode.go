package audio

import (
	"context"
	"fmt"
	"io"
	"math"

	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/telemetry"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*GainNode)(nil)

type GainNode struct {
	logger *logging.Logger
	err    error

	factor float32
}

func (node *GainNode) Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) {
	ctx, span := telemetry.Tracer.Start(ctx, "GainNode.Tick")
	defer span.End()

	node.err = nil

	if len(ins) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
		return
	}

	if len(outs) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 1, 1, len(outs))
		return
	}

	bytes, err := io.ReadAll(ins[0])
	telemetry.Logger.DebugContext(ctx, "GainNode copied data from input to internal buffer", "n", len(bytes), "error", err)

	if err != nil {
		node.err = fmt.Errorf("failed to copy data from input to internal buffer: %w", err)
		return
	}

	stream := codecs.BytesToS16LE(bytes)

	for i, sample := range stream {
		f32 := float32(sample) * node.factor
		switch {
		case f32 < math.MinInt16: // underflow so set the sample to the min value
			stream[i] = math.MinInt16
		case f32 > math.MaxInt16: // overflow so set the sample to the max value
			stream[i] = math.MaxInt16
		default:
			stream[i] = int16(f32)
		}
	}

	n, err := outs[0].Write(codecs.S16LEToBytes(stream))
	telemetry.Logger.DebugContext(ctx, "GainNode copied data from internal buffer to output", "n", n, "error", err)

	if err != nil {
		node.err = fmt.Errorf("failed to copy data from internal buffer to output: %w", err)
		return
	}
}

func (node *GainNode) Err() error {
	return node.err
}

func NewGainNode(logger *logging.Logger, factor float32) *GainNode {
	return &GainNode{logger: logger, factor: factor}
}
