package graph

import (
	"bytes"
	"fmt"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

type InvalidConnectionConfigErr error

func newInvalidConnectionConfigErr(nMin, nMax, nActual int) InvalidConnectionConfigErr {
	return fmt.Errorf("invalid audio graph connection configuration. node requires min=%d, max=%d, but got actual=%d", nMin, nMax, nActual)
}

// a unit that processess all input channels are writes to all output channels.
// once Start()ed, the node should not stop processing under any condition unless Stop() is called.
type Node interface {
	// process inputs and writing them to outputs
	// this function should block
	Tick(ins []io.Reader, outs []io.Writer) error
}

type Connection struct {
	bytes.Buffer

	from Node
	to   Node
}

type Graph struct {
	*CompositeNode
}

func (graph *Graph) Tick() error {
	return graph.CompositeNode.Tick([]io.Reader{}, []io.Writer{})
}

func NewGraph(logger *logging.Logger) *Graph {
	return &Graph{NewCompositeNode(logger)}
}
