package parallelgraph

import (
	"fmt"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*WriterNode)(nil)

type WriterNode struct {
	logger *logging.Logger
	w      io.Writer
	ins    []<-chan byte
	outs   []chan<- byte
	errs   chan error
	stop   chan FlushPolicy
}

func (node *WriterNode) Start() error {
	if len(node.ins) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(node.ins))
	}

	if len(node.outs) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(node.outs))
	}

	go node.process()

	return nil
}

func (node *WriterNode) Stop(flush FlushPolicy) error {
	defer func() {
		close(node.errs)
		node.logger.Debug("closed Errors() channel")
	}()

	node.stop <- flush

	return nil
}

func (node *WriterNode) Errors() <-chan error {
	return node.errs
}

func (node *WriterNode) addInput(in <-chan byte) {
	node.ins = append(node.ins, in)
}

func (node *WriterNode) addOutput(out chan<- byte) {
	node.outs = append(node.outs, out)
}

func (node *WriterNode) process() {
	buf := [1]byte{}

	for {
		select {
		case flush := <-node.stop:
			node.logger.Debug("received done signal")
			_ = flush // TODO: handle flush policy
			break
		case buf[0] = <-node.ins[0]:
			_, err := node.w.Write(buf[:])
			//node.logger.Debug("wrote bytes to writer", "n", n, "error", err)

			if err != nil {
				node.errs <- fmt.Errorf("error while writing bytes to writer: %w", err)

			}
		}
	}
}

func NewWriterNode(logger *logging.Logger, w io.Writer) *WriterNode {
	return &WriterNode{logger: logger, w: w, errs: make(chan error, 1)}
}
