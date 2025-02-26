package graph

import (
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	logger *logging.Logger

	w io.Writer
}

func (node *WriterNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(ins))
	}

	if len(outs) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(outs))
	}

	n, err := io.Copy(node.w, ins[0])
	node.logger.Debug("WriterNode copied data from input to writer", "n", n, "error", err)

	return err
}

func NewWriterNode(logger *logging.Logger, w io.Writer) *WriterNode {
	return &WriterNode{logger: logger, w: w}
}
