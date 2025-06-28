package audio

import (
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	logger *logging.Logger

	r    io.Reader
	size int64
	err  error
}

func (node *ReaderNode) Tick(ins []io.Reader, outs []io.Writer) {
	node.err = nil

	if len(ins) != 0 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 0, 0, len(ins))
		return
	}

	if len(outs) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 1, 1, len(outs))
		return
	}

	var n int64
	n, node.err = io.CopyN(outs[0], node.r, node.size)
	node.logger.Debug("ReaderNode copied data from reader to output", "n", n, "error", node.err)
}

func (node *ReaderNode) Err() error {
	return node.err
}

func NewReaderNode(logger *logging.Logger, r io.Reader, size int64) *ReaderNode {
	return &ReaderNode{logger: logger, r: r, size: size, err: nil}
}
