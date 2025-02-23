package parallelgraph

import (
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

const ConnectionBufferSize = 0x8000

type FlushPolicy byte

const (
	FlushPolicy_NoFlush = iota
	FlushPolicy_Flush
)

type InvalidConnectionConfigErr error

func newInvalidConnectionConfigErr(nMin, nMax, nActual int) InvalidConnectionConfigErr {
	return fmt.Errorf("invalid audio graph connection configuration. node requires min=%d, max=%d, but got actual=%d", nMin, nMax, nActual)
}

// a unit that processess all input channels are writes to all output channels.
// once Start()ed, the node should not stop processing under any condition unless Stop() is called.
type Node interface {
	// start processing inputs and writing them to outputs
	// this function should not block.
	Start() error

	// stop processing inputs and writing outputs
	// this function should not block.
	Stop(flush FlushPolicy) error

	// errors generated by the node.
	// channel is closed when Stop() is called.
	Errors() <-chan error

	addInput(in <-chan byte)
	addOutput(out chan<- byte)
}

type Connection struct {
	from Node
	to   Node

	buf chan byte
}

type Graph struct {
	*CompositeNode
}

func NewGraph(logger *logging.Logger) *Graph {
	return &Graph{NewCompositeNode(logger)}
}
