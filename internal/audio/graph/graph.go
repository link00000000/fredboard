package graph

import (
	"fmt"
	"io"
)

type AudioGraph struct {
	internal *CompositeNode
}

func (graph *AudioGraph) AddNode(n AudioGraphNode) error {
	err := graph.internal.AddNode(n)

	if err != nil {
		return fmt.Errorf("AudioGraphNode.AddNode() error while adding node to interal graph: %w", err)
	}

	return nil
}

func (graph *AudioGraph) RemoveNode(node AudioGraphNode) error {
	err := graph.internal.RemoveNode(node)

	if err != nil {
		return fmt.Errorf("AudioGraphNode.RemoveNode() error while removing node from interal graph: %w", err)
	}

	return nil
}

func (graph *AudioGraph) CreateConnection(from, to AudioGraphNode) error {
	err := graph.internal.CreateConnection(from, to)

	if err != nil {
		return fmt.Errorf("AudioGraph.CreateConnection error: %w", err)
	}

	return nil
}

func (graph *AudioGraph) DestroyConnection(from, to AudioGraphNode) error {
	err := graph.internal.DestroyConnection(from, to)

	if err != nil {
		return fmt.Errorf("AudioGraph.RemoveConnection error: %w", err)
	}

	return nil
}

func (graph *AudioGraph) Tick() error {
	err := graph.internal.Tick([]io.Reader{}, []io.Writer{})

	if err != nil {
		return fmt.Errorf("AudioGraph.Tick error: %w", err)
	}

	return nil
}

func NewAudioGraph() *AudioGraph {
	return &AudioGraph{
		internal: NewCompositeNode(),
	}
}
