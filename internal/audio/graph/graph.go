package graph

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph/nodes"
)

var (
	ErrGraphAlreadyContainsNode       = errors.New("graph already contains node")
	ErrGraphAlreadyContainsConnection = errors.New("graph already contains connection")
	ErrNotExist                       = errors.New("not exists")
	ErrInvalidNOuts                   = errors.New("invalid number of outs")
)

type AudioGraphConnectionBuffer struct {
	buf []byte
}

type AudioGraphConnection struct {
	from nodes.AudioGraphNode
	to   nodes.AudioGraphNode
	buf  bytes.Buffer
}

// Implements [io.Reader]
func (conn *AudioGraphConnection) Read(p []byte) (int, error) {
	return conn.buf.Read(p)
}

// Implements [io.Writer]
func (conn *AudioGraphConnection) Write(p []byte) (int, error) {
	return conn.buf.Write(p)
}

func NewAudioGraphConnection(from, to nodes.AudioGraphNode) *AudioGraphConnection {
	return &AudioGraphConnection{from: from, to: to}
}

type AudioGraph struct {
	nodes       []nodes.AudioGraphNode
	connections []*AudioGraphConnection
}

func (graph *AudioGraph) AddNode(node nodes.AudioGraphNode) error {
	// TODO: Make a fast version that does not do this check
	if slices.Contains(graph.nodes, node) {
		return ErrGraphAlreadyContainsNode
	}

	graph.nodes = append(graph.nodes, node)

	return nil
}

// TODO: RemoveNode()

func (graph *AudioGraph) CreateConnection(from, to nodes.AudioGraphNode) error {
	// TODO: Make a fast version that does not do this check
	idx := slices.IndexFunc(graph.connections, func(conn *AudioGraphConnection) bool {
		return conn.from == from && conn.to == to
	})

	if idx != -1 {
		return ErrGraphAlreadyContainsConnection
	}

	graph.connections = append(graph.connections, NewAudioGraphConnection(from, to))

	return nil
}

func (graph *AudioGraph) RemoveConnection(from, to nodes.AudioGraphNode) error {
	del := slices.DeleteFunc(graph.connections, func(conn *AudioGraphConnection) bool {
		return conn.from == from && conn.to == to
	})

	if len(del) == 0 {
		return ErrNotExist
	}

	return nil
}

// TODO: Cancel with context
func (graph *AudioGraph) Tick() error {
	for _, node := range graph.nodes {
		err := node.PreTick()
		if err != nil {
			return fmt.Errorf("AudioGraph.ProcessNextSample error: %w", err)
		}
	}

	leafNodes := make([]nodes.AudioGraphNode, 0, len(graph.connections))
	for _, node := range graph.nodes {
		if !slices.ContainsFunc(graph.connections, func(conn *AudioGraphConnection) bool { return conn.from == node }) {
			leafNodes = append(leafNodes, node)
		}
	}

	for _, node := range leafNodes {
		conns := make([]*AudioGraphConnection, 0, len(graph.connections))
		for _, conn := range graph.connections {
			if conn.to == node {
				conns = append(conns, conn)
			}
		}

		err := graph.TickNode(node)
		if err != nil {
			return fmt.Errorf("AudioGraph.ProcessNextSample error: %w", err)
		}
	}

	for _, node := range graph.nodes {
		err := node.PostTick()
		if err != nil {
			return fmt.Errorf("AudioGraph.ProcessNextSample error: %w", err)
		}
	}

	return nil
}

func (graph *AudioGraph) TickNode(node nodes.AudioGraphNode) error {
	dependencies := make([]io.Reader, 0)
	dependents := make([]io.Writer, 0)
	for _, conn := range graph.connections {
		if conn.to == node {
			dependencies = append(dependencies, conn)

			// Process dependencies
			err := graph.TickNode(conn.from)
			if err != nil {
				return fmt.Errorf("ProcessSample errorr: %w", err)
			}
		}

		if conn.from == node {
			dependents = append(dependents, conn)
		}
	}

	err := node.Tick(dependencies, dependents)
	if err != nil {
		return fmt.Errorf("ProcessSample error: %w", err)
	}

	return nil
}

func NewAudioGraph() *AudioGraph {
	return &AudioGraph{
		nodes:       make([]nodes.AudioGraphNode, 0),
		connections: make([]*AudioGraphConnection, 0),
	}
}
