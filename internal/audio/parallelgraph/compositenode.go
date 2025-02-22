package parallelgraph

import (
	"context"
	"fmt"
	"io"
	"sync"
)

var _ Node = (*CompositeNode)(nil)

type CompositeNode struct {
	childNodes  []Node
	connections []*Connection

	errs chan error
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
			defer wg.Done()

			for err := range childNode.Errors() {
				node.errs <- err
			}
		}()

		err := childNode.Start(ctx, childIns, childOuts)
		if err != nil {
			return fmt.Errorf("failed to start child node: %w", err)
		}
	}

	go func() {
		defer close(node.errs)

		wg.Wait()
	}()

	return nil
}

func (node *CompositeNode) Errors() <-chan error {
	return node.errs
}

func (node *CompositeNode) AddNode(n Node) error {
	// TODO: Validation

	node.childNodes = append(node.childNodes, n)

	return nil
}

func (node *CompositeNode) CreateConnection(from, to Node) error {
	// TODO: Validation

	node.connections = append(node.connections, &Connection{from: from, to: to})

	return nil
}

func NewCompositeNode() *CompositeNode {
	return &CompositeNode{errs: make(chan error, 1)}
}
