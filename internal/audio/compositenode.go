package audio

import (
	"io"
	"iter"
	"slices"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*CompositeNode)(nil)

type CompositeNode struct {
	logger *logging.Logger

	childNodes    []Node
	connections   []*Connection
	input, output Node
}

func (node *CompositeNode) Tick(ins []io.Reader, outs []io.Writer) {
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

	for _, n := range queue {
		nins := make([]io.Reader, 0)
		nouts := make([]io.Writer, 0)

		if n == node.input {
			nins = append(nins, ins...)
		}

		if n == node.output {
			nouts = append(nouts, outs...)
		}

		for _, conn := range node.connections {
			if conn.to == n {
				nins = append(nins, conn)
			}

			if conn.from == n {
				nouts = append(nouts, conn)
			}
		}

		n.Tick(nins, nouts)
	}
}

func (node *CompositeNode) Err() error {
	return nil
}

func (node *CompositeNode) AddNode(n Node) {
	// TODO: assert that n is a pointer
	// TODO: assert that n is not already in the graph

	node.childNodes = append(node.childNodes, n)
}

func (node *CompositeNode) RemoveNode(n Node) {
	// TODO: assert that n is a pointer
	// TODO: assert that n is in the graph

	node.connections = slices.DeleteFunc(node.connections, func(c *Connection) bool { return c.from == n || c.to == n })
	node.childNodes = slices.DeleteFunc(node.childNodes, func(nn Node) bool { return n == nn })
}

func (node *CompositeNode) Nodes() iter.Seq2[int, Node] {
	return func(yield func(int, Node) bool) {
		for i, n := range node.childNodes {
			if !yield(i, n) {
				return
			}
		}
	}
}

func (node *CompositeNode) CreateConnection(from, to Node) {
	// TODO: assert that to and from are pointers
	// TODO: assert that to and from are in the graph
	// TODO: assert that the connection does not already exist
	// TODO: assert that cycle isn't created

	node.connections = append(node.connections, &Connection{from: from, to: to})
}

func (node *CompositeNode) RemoveConnection(from, to Node) {
	// TODO: assert that to and from are pointers
	// TODO: assert that to and from are in the graph
	// TODO: assert that the connection exists

	node.connections = slices.DeleteFunc(node.connections, func(conn *Connection) bool { return conn.from == from && conn.to == to })
}

func (node *CompositeNode) SetAsInput(n Node) {
	// TODO: assert that n is a pointer
	// TODO: assert that n is in the graph

	node.input = n
}

func (node *CompositeNode) SetAsOutput(n Node) {
	// TODO: assert that n is a pointer
	// TODO: assert that n is in the graph

	node.output = n
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
	}
}
