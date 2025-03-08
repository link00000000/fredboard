package audio

import (
	"errors"
	"fmt"
	"io"
	"math"
	"slices"

	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*MixerNode)(nil)

type MixerNode struct {
	logger *logging.Logger
	err    error
}

func (node *MixerNode) Tick(ins []io.Reader, outs []io.Writer) {
	node.err = nil

	if len(ins) <= 0 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 0, connection_Unbounded, len(ins))
		return
	}

	if len(outs) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 1, 1, len(outs))
		return
	}

	errs := make([]error, 0)
	mixedStream := make([]int16, 0) // TODO: cache

	for inIdx, in := range ins {
		bytes, err := io.ReadAll(in)
		node.logger.Debug("MixerNode copied data from input to internal buffer", "input", inIdx, "n", len(bytes), "error", err)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read from input %d: %w", inIdx, err))
			continue
		}

		stream := codecs.BytesToS16LE(bytes)

		if len(stream) > len(mixedStream) {
			node.logger.Debug("mixedStream buffer is too small. growing buffer", "oldcap", "newcap")

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
	node.logger.Debug("MixerNode copied data from internal buffer to output", "n", n, "error", err)

	if err != nil {
		errs = append(errs, fmt.Errorf("failed to copy data from internal buffer to output: %w", err))
	}

	node.err = errors.Join(errs...)
}

func (node *MixerNode) Err() error {
	return node.err
}

func NewMixerNode(logger *logging.Logger) *MixerNode {
	return &MixerNode{logger: logger, err: nil}
}
