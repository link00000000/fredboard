package parallelgraph

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*CompositeNode)(nil)

type CompositeNode struct {
	logger *logging.Logger
	errs   chan error

	childNodes  []Node
	connections []*Connection
}

func (node *CompositeNode) Start(ctx context.Context, ins []io.Reader, outs []io.Writer) error {
	if !(len(ins) < 2) {
		return newInvalidConnectionConfigErr(0, 1, len(ins))
	}

	if !(len(outs) < 2) {
		return newInvalidConnectionConfigErr(0, 1, len(outs))
	}

	// waits for all child nodes to close their error channels
	var wg sync.WaitGroup

	for _, childNode := range node.childNodes {
		childIns := make([]io.Reader, 0)
		childOuts := make([]io.Writer, 0)

		for _, conn := range node.connections {
			if conn.to == childNode {
				childIns = append(childIns, conn)
			}

			if conn.from == childNode {
				childOuts = append(childOuts, conn)
			}
		}

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				node.logger.Debug("child node completed", "child", childNode)
			}()

			node.logger.Debug("waiting for child node to complete", "child", childNode)

			for err := range childNode.Errors() {
				node.errs <- err
			}
		}()

		node.logger.Debug("starting child node", "child", childNode)
		err := childNode.Start(ctx, childIns, childOuts)
		if err != nil {
			return fmt.Errorf("failed to start child node: %w", err)
		}
	}

	go func() {
		defer func() {
			close(node.errs)
			node.logger.Debug("closed Errors() channel")
		}()

		node.logger.Debug("waiting for all child nodes")
		wg.Wait()

		node.logger.Debug("all child nodes complete")
	}()

	return nil
}

func (node *CompositeNode) Errors() <-chan error {
	return node.errs
}

func (node *CompositeNode) AddNode(n Node) error {
	// TODO: Validation

	node.childNodes = append(node.childNodes, n)
	node.logger.Debug("added child node", "child", n)

	return nil
}

func (node *CompositeNode) CreateConnection(from, to Node) error {
	// TODO: Validation

	node.connections = append(node.connections, &Connection{from: from, to: to})
	node.logger.Debug("created graph connection", "from", from, "to", to)

	return nil
}

func NewCompositeNode(logger *logging.Logger) *CompositeNode {
	return &CompositeNode{logger: logger, errs: make(chan error, 1)}
}

func (node *CompositeNode) FlushAndStop() {
	flushing := make([]Node, 0)

	var flushNodeAndParents func(n Node)
	flushNodeAndParents = func(n Node) {
		if slices.Contains(flushing, n) {
			// this node and its parents are already being flushed by some other dependency
			// dont flush it ands its dependencies again
			return
		}

		// flush all dependencies (and their dependencies) before flushing this node
		for _, parent := range node.findParentNodes(n) {
			flushNodeAndParents(parent)
		}

		n.FlushAndStop()
	}

	for _, leaf := range node.findLeafNodes() {
		flushNodeAndParents(leaf)
	}
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

func (node *CompositeNode) findParentNodes(child Node) []Node {
	parents := make([]Node, 0)

	for _, conn := range node.connections {
		if conn.to == child {
			parents = append(parents, conn.from)
		}
	}

	return parents
}
