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
		return fmt.Errorf("AudioGraph.AddNode error: %w", err)
	}

	return nil
}

// TODO: RemoveNode

func (graph *AudioGraph) CreateConnection(from, to AudioGraphNode) error {
	err := graph.internal.CreateConnection(from, to)

	if err != nil {
		return fmt.Errorf("AudioGraph.CreateConnection error: %w", err)
	}

	return nil
}

func (graph *AudioGraph) RemoveConnection(from, to AudioGraphNode) error {
	err := graph.internal.RemoveConnection(from, to)

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
