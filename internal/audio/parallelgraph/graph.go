package parallelgraph

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

type InvalidConnectionConfigErr error

func newInvalidConnectionConfigErr(nMin, nMax, nActual int) InvalidConnectionConfigErr {
	return fmt.Errorf("invalid audio graph connection configuration. node requires min=%d, max=%d, but got actual=%d", nMin, nMax, nActual)
}

type Node interface {
	// ins and outs will never be nil, but may be an empty slice
	Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error

	// sends errors generated from the node. this channel is closed when the node is complete
	// and is a good place to wait for the node
	Errors() <-chan error
}

type Connection struct {
	bytes.Buffer

	from Node
	to   Node
}

type Graph struct {
	*CompositeNode
}

func (graph *Graph) Start(ctx context.Context) error {
	return graph.CompositeNode.Start(ctx, []io.Reader{}, []io.Writer{})
}

func NewGraph() *Graph {
	return &Graph{NewCompositeNode()}
}
