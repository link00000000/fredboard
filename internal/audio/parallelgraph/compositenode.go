package parallelgraph

import (
	"accidentallycoded.com/fredboard/v3/internal/errors"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*CompositeNode)(nil)

type CompositeNode struct {
	logger *logging.Logger
	ins    []<-chan byte
	outs   []chan<- byte
	errs   chan error

	childNodes  []Node
	connections []*Connection
}

func (node *CompositeNode) Start() error {
	if !(len(node.ins) < 2) {
		return newInvalidConnectionConfigErr(0, 1, len(node.ins))
	}

	if !(len(node.outs) < 2) {
		return newInvalidConnectionConfigErr(0, 1, len(node.outs))
	}

	errs := errors.NewErrorList()
	for _, childNode := range node.childNodes {
		errs.Add(childNode.Start())
	}

	return errs.Join()
}

func (node *CompositeNode) Stop(flush FlushPolicy) error {
	defer func() {
		close(node.errs)
		node.logger.Debug("closed Errors() channel")
	}()

	errs := errors.NewErrorList()
	for _, childNode := range node.childNodes {
		errs.Add(childNode.Stop(flush))
	}

	return errs.Join()
}

func (node *CompositeNode) Errors() <-chan error {
	return node.errs
}

func (node *CompositeNode) AddNode(n Node) error {
	// TODO: validate that the node does not already exist in the graph

	node.childNodes = append(node.childNodes, n)
	node.logger.Debug("added child node", "child", n)

	return nil
}

func (node *CompositeNode) CreateConnection(from, to Node) error {
	// TODO: validate that the connection does not already exist
	// TODO: validate that all of the nodes are in the graph already
	// TODO: validate that there are no cycles in the graph

	conn := &Connection{from: from, to: to, buf: make(chan byte, ConnectionBufferSize)}
	node.connections = append(node.connections, conn)
	node.logger.Debug("created graph connection", "from", from, "to", to)

	from.addOutput(conn.buf)
	to.addInput(conn.buf)

	return nil
}

func (node *CompositeNode) addInput(in <-chan byte) {
	node.ins = append(node.ins, in)
}

func (node *CompositeNode) addOutput(out chan<- byte) {
	node.outs = append(node.outs, out)
}

func NewCompositeNode(logger *logging.Logger) *CompositeNode {
	return &CompositeNode{
		logger: logger,
		ins:    make([]<-chan byte, 0),
		outs:   make([]chan<- byte, 0),
		errs:   make(chan error, 1),

		childNodes:  make([]Node, 0),
		connections: make([]*Connection, 0),
	}
}
