package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"

	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

const (
	connection_Unbounded = -1
	connectionType_In    = "in"
	connectionType_Out   = "out"
)

type InvalidConnectionConfigErr error

func newInvalidConnectionConfigErr(node Node, connectionType string, nMin, nMax, nActual int) InvalidConnectionConfigErr {
	sMin := fmt.Sprintf("%d", nMin)
	if nMin == connection_Unbounded {
		sMin = "UNBOUNDED"
	}

	sMax := fmt.Sprintf("%d", nMax)
	if nMax == connection_Unbounded {
		sMax = "UNBOUNDED"
	}

	return fmt.Errorf("invalid audio graph connection configuration. %s requires min=%s, max=%s \"%s\" connections, but got actual=%d", reflect.TypeOf(node).String(), sMin, sMax, connectionType, nActual)
}

// a unit that processess all input channels and writes to all output channels.
type Node interface {
	// process inputs and writing them to outputs
	// this function should block until all work is complete, returning an error if any
	Tick(ctx context.Context, ins []io.Reader, outs []io.Writer) (err error)
}

type Connection struct {
	bytes.Buffer

	from Node
	to   Node
}

type Graph struct {
	compositeNode *CompositeNode
}

func (graph *Graph) Tick(ctx context.Context) (err error) {
	ctx, span := telemetry.Tracer.Start(ctx, "audio.Graph.Tick")
	defer span.End()

	return graph.compositeNode.Tick(ctx, []io.Reader{}, []io.Writer{})
}

func (graph *Graph) AddNode(n Node) {
	graph.compositeNode.AddNode(n)
}

func (graph *Graph) RemoveNode(n Node) {
	graph.compositeNode.RemoveNode(n)
}

func (graph *Graph) CreateConnection(from, to Node) {
	graph.compositeNode.CreateConnection(from, to)
}

func (graph *Graph) RemoveConnection(from, to Node) {
	graph.compositeNode.RemoveConnection(from, to)
}

func NewGraph() *Graph {
	return &Graph{NewCompositeNode()}
}
