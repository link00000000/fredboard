package audio

import (
	"fmt"
	"io"

	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*TeeNode)(nil)

type TeeNode struct {
	logger *logging.Logger
	err    error
}

func (node *TeeNode) Tick(ins []io.Reader, outs []io.Writer) {
	node.err = nil

	if len(ins) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
		return
	}

	if len(outs) <= 0 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 0, connection_Unbounded, len(outs))
		return
	}

	bytes, err := io.ReadAll(ins[0])
	node.logger.Debug("TeeNode copied data from input to internal buffer", "n", len(bytes), "error", err)

	if err != nil {
		node.err = fmt.Errorf("failed to read from input: %w", err)
		return
	}

	errs := make([]error, 0)

	for outIdx, out := range outs {
		n, err := out.Write(bytes)
		node.logger.Debug("TeeNode copied data from input to internal buffer", "n", n, "error", err)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to write data to output %d: %w", outIdx, err))
			continue
		}
	}
}

func (node *TeeNode) Err() error {
	return node.err
}

func NewTeeNode(logger *logging.Logger) *TeeNode {
	return &TeeNode{logger: logger}
}
