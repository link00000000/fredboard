package audio

import (
	"io"

	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	logger *logging.Logger

	w   io.Writer
	err error
}

func (node *WriterNode) Tick(ins []io.Reader, outs []io.Writer) {
	node.err = nil

	if len(ins) != 1 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_In, 1, 1, len(ins))
		return
	}

	if len(outs) != 0 {
		node.err = newInvalidConnectionConfigErr(node, connectionType_Out, 0, 0, len(outs))
		return
	}

	var n int64
	n, node.err = io.Copy(node.w, ins[0])
	node.logger.Debug("WriterNode copied data from input to writer", "n", n, "error", node.err)
}

func (node *WriterNode) Err() error {
	return node.err
}

func NewWriterNode(logger *logging.Logger, w io.Writer) *WriterNode {
	return &WriterNode{logger: logger, w: w, err: nil}
}
