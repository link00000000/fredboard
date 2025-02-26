package graph

import (
	"errors"
	"io"
	"slices"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*CompositeNode)(nil)

type CompositeNode struct {
	logger *logging.Logger

	childNodes  []Node
	connections []*Connection
	err         error
}

func (node *CompositeNode) Tick(ins []io.Reader, outs []io.Writer) {
	node.err = nil

	queue := make([]Node, 0)

	var enqueue func(n Node)
	enqueue = func(n Node) {
		parents := node.findParentsOf(n)
		for _, parent := range parents {
			if !slices.Contains(queue, parent) {
				enqueue(parent)
			}
		}

		queue = append(queue, n)
	}

	leaves := node.findLeafNodes()

	for _, leaf := range leaves {
		enqueue(leaf)
	}

	errs := make([]error, 0)
	for _, n := range queue {
		ins := make([]io.Reader, 0)
		outs := make([]io.Writer, 0)
		for _, conn := range node.connections {
			if conn.to == n {
				ins = append(ins, conn)
			}

			if conn.from == n {
				outs = append(outs, conn)
			}
		}

		n.Tick(ins, outs)
		errs = append(errs, n.Err())
	}

	node.err = errors.Join(errs...)
}

func (node *CompositeNode) Err() error {
	return node.err
}

func (node *CompositeNode) AddNode(n Node) {
	// TODO: assert that n is a pointer
	// TODO: assert that n is not already in the graph

	node.childNodes = append(node.childNodes, n)
}

func (node *CompositeNode) CreateConnection(from, to Node) {
	// TODO: assert that to and from are pointers
	// TODO: assert that to and from are in the graph
	// TODO: assert that the connection does not already exist
	// TODO: assert that cycle isn't created

	node.connections = append(node.connections, &Connection{from: from, to: to})
}

// finds all nodes that are not a 'from' in any connection
func (node *CompositeNode) findLeafNodes() []Node {
	leaves := make([]Node, 0)

	for _, childNode := range node.childNodes {
		isLeaf := !slices.ContainsFunc(node.connections, func(conn *Connection) bool {
			return conn.from == childNode
		})

		if isLeaf {
			leaves = append(leaves, childNode)
		}
	}

	return leaves
}

func (node *CompositeNode) findParentsOf(child Node) []Node {
	parents := make([]Node, 0)

	for _, conn := range node.connections {
		if conn.to == child {
			parents = append(parents, conn.from)
		}
	}

	return parents
}

func NewCompositeNode(logger *logging.Logger) *CompositeNode {
	return &CompositeNode{
		logger:      logger,
		childNodes:  make([]Node, 0),
		connections: make([]*Connection, 0),
		err:         nil,
	}
}
