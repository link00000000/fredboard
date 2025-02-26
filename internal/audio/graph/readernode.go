package graph

import (
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	logger *logging.Logger

	r    io.Reader
	size int64
}

func (node *ReaderNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if len(ins) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(ins))
	}

	if len(outs) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(outs))
	}

	n, err := io.CopyN(outs[0], node.r, node.size)
	node.logger.Debug("ReaderNode copied data from reader to output", "n", n, "error", err)

	return err
}

func NewReaderNode(logger *logging.Logger, r io.Reader, size int64) *ReaderNode {
	return &ReaderNode{logger: logger, r: r, size: size}
}
