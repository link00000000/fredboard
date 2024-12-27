package graph

import (
	"errors"
	"fmt"
	"io"
	"log"
)

type NodeState byte

const (
	NodeState_Ready NodeState = iota
	NodeState_Running
	NodeState_Done
	NodeState_Stopped
)

type GraphNode interface {
	io.ReadWriter

	Start() error
	Stop() error
	Wait()

	State() NodeState

	GetParentNodes() []GraphNode
	GetChildNodes() []GraphNode

	AddParent(parent GraphNode)
	AddChild(child GraphNode)

	notifyAddedAsParentOf(child GraphNode)
	notifyAddedAsChildOf(parent GraphNode)
}

// Connects an input and output node without performing transformations.
type PassthroughNode struct {
	in  GraphNode
	out GraphNode
}

func (passthrough *PassthroughNode) Read(p []byte) (int, error) {
	if passthrough.in == nil {
		panic("cannot read from PassthroughNode that does not have an 'in' node")
	}

	return passthrough.in.Read(p)
}

func (passthrough *PassthroughNode) Write(p []byte) (int, error) {
	if passthrough.out == nil {
		panic("cannot write to PassthroughNode that does not have an 'out' node")
	}

	return passthrough.out.Write(p)
}

func (passthrough *PassthroughNode) Start() error {
	if passthrough.in == nil {
		panic("cannot start PassthroughNode because 'in' node is nil")
	}

	return passthrough.in.Start()
}

func (passthrough *PassthroughNode) Stop() error {
	if passthrough.in == nil {
		panic("cannot stop PassthroughNode because 'in' node is nil")
	}

	return passthrough.in.Start()
}

func (passthrough *PassthroughNode) Wait() {
	if passthrough.in == nil {
		panic("cannot wait PassthroughNode because 'in' node is nil")
	}

	passthrough.in.Wait()
}

func (passthrough *PassthroughNode) State() NodeState {
	if passthrough.in == nil {
		panic("cannot get state of PassthroughNode because 'in' node is nil")
	}

	return passthrough.in.State()
}

func (passthrough *PassthroughNode) GetParentNodes() []GraphNode {
	if passthrough.in == nil {
		return []GraphNode{}
	}

	return []GraphNode{passthrough.in}
}

func (passthrough *PassthroughNode) GetChildNodes() []GraphNode {
	if passthrough.out == nil {
		return []GraphNode{}
	}

	return []GraphNode{passthrough.out}
}

func (passthrough *PassthroughNode) AddParent(parent GraphNode) {
	if passthrough.in != nil {
		panic("PassthroughNode cannot have more than 1 parent")
	}

	passthrough.in = parent
	parent.notifyAddedAsParentOf(passthrough)
}

func (passthrough *PassthroughNode) AddChild(child GraphNode) {
	if passthrough.out != nil {
		panic("PassthroughNode cannot have more than 1 child")
	}

	passthrough.out = child
	child.notifyAddedAsChildOf(passthrough)
}

func (passthrough *PassthroughNode) notifyAddedAsParentOf(child GraphNode) {
	passthrough.out = child
}

func (passthrough *PassthroughNode) notifyAddedAsChildOf(parent GraphNode) {
	passthrough.in = parent
}

func NewPassthroughNode() *PassthroughNode {
	return &PassthroughNode{}
}

// Contains a subgraph that exposes the same API as a single node.
// A composite node as 0 or 1 in nodes and 0 or 1 out nodes.
type CompositeNode struct {
	in  GraphNode
	out GraphNode

	// Bottom-most consumer nodes of the subgraph
	leafNodes []GraphNode
}

func (composite *CompositeNode) Read(p []byte) (int, error) {
	if composite.out == nil {
		panic("cannot read from a CompositeNode that does not have an 'out' node")
	}

	return composite.out.Read(p)
}

func (composite *CompositeNode) Write(p []byte) (int, error) {
	if composite.in == nil {
		panic("cannot read from CompositeNode that does not have an 'out' node")
	}

	return composite.in.Write(p)
}

func (composite *CompositeNode) Start() error {
	errs := make([]error, 0)

	for _, node := range composite.leafNodes {
		err := node.Start()

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("CompositeNode.Start(): %w", errors.Join(errs...))
	}

	return nil
}

func (composite *CompositeNode) Stop() error {
	errs := make([]error, 0)

	for _, node := range composite.leafNodes {
		err := node.Stop()

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("CompositeNode.Stop(): %w", errors.Join(errs...))
	}

	return nil
}

func (composite *CompositeNode) Wait() {
	for _, node := range composite.leafNodes {
		node.Wait()
	}
}

// If any leaf node is [NodeState_Running], the state of the subgraph is [NodeState_Running]
// Else, if any leaf node is [NodeState_Stopped], the state of the subgraph is [NodeState_Stopped]
// Else, if any leaf node is [NodeState_Done], the state of the subgraph is [NodeState_Done]
// Else, all leaf nodes have the state [NodeState_Ready] and the state of the subgraph is [NodeState_Ready]
func (compsite *CompositeNode) State() NodeState {
	activeStatesInSubgraph := make(map[NodeState]bool)
	for _, node := range compsite.leafNodes {
		activeStatesInSubgraph[node.State()] = true
	}

	if _, ok := activeStatesInSubgraph[NodeState_Running]; ok {
		return NodeState_Running
	}

	if _, ok := activeStatesInSubgraph[NodeState_Stopped]; ok {
		return NodeState_Stopped
	}

	if _, ok := activeStatesInSubgraph[NodeState_Done]; ok {
		return NodeState_Done
	}

	return NodeState_Running
}

func (composite *CompositeNode) GetParentNodes() []GraphNode {
	if composite.in == nil {
		return []GraphNode{}
	}

	return []GraphNode{composite.in}
}

func (composite *CompositeNode) GetChildNodes() []GraphNode {
	if composite.out == nil {
		return []GraphNode{}
	}

	return []GraphNode{composite.out}
}

func (composite *CompositeNode) AddParent(parent GraphNode) {
	if composite.in != nil {
		panic("CompositeNode cannot have more than 1 parent")
	}

	composite.in = parent
	parent.notifyAddedAsParentOf(composite)
}

func (composite *CompositeNode) AddChild(child GraphNode) {
	if composite.out != nil {
		panic("CompositeNode cannot have more than 1 child")
	}

	composite.out = child
	child.notifyAddedAsChildOf(composite)
}

func (composite *CompositeNode) notifyAddedAsParentOf(child GraphNode) {
	composite.out = child
}

func (composite *CompositeNode) notifyAddedAsChildOf(parent GraphNode) {
	composite.in = parent
}

func NewCompositeNode() *CompositeNode {
	return &CompositeNode{
		leafNodes: make([]GraphNode, 0),
	}
}

// Reads zeroes until stopped
type ZeroSourceNode struct {
	out GraphNode

	state NodeState
	stop  chan struct{}
}

func (zeroSource *ZeroSourceNode) Read(p []byte) (int, error) {
	for i := range len(p) {
		p[i] = 0x00
	}

	return len(p), nil
}

func (zeroSource *ZeroSourceNode) Write(p []byte) (int, error) {
	panic("cannot write to ZeroSourceNode")
}

func (zeroSource *ZeroSourceNode) Start() error {
	zeroSource.state = NodeState_Running

	return nil
}

func (zeroSource *ZeroSourceNode) Stop() error {
	zeroSource.state = NodeState_Stopped
	zeroSource.stop <- struct{}{}

	return nil
}

func (zeroSource *ZeroSourceNode) Wait() {
	<-zeroSource.stop
}

func (zeroSource *ZeroSourceNode) State() NodeState {
	return zeroSource.state
}

func (zeroSource *ZeroSourceNode) GetParentNodes() []GraphNode {
	return []GraphNode{}
}

func (zeroSource *ZeroSourceNode) GetChildNodes() []GraphNode {
	if zeroSource.out == nil {
		return []GraphNode{}
	}

	return []GraphNode{zeroSource.out}
}

func (zeroSource *ZeroSourceNode) AddParent(parent GraphNode) {
	panic("cannot add parent to ZeroSourceNode")
}

func (zeroSource *ZeroSourceNode) AddChild(child GraphNode) {
	if zeroSource.out != nil {
		panic("ZeroSourceNode cannot have more than 1 child")
	}

	zeroSource.out = child
	child.notifyAddedAsChildOf(zeroSource)
}

func (zeroSource *ZeroSourceNode) notifyAddedAsParentOf(child GraphNode) {
	zeroSource.out = child
}

func (zeroSource *ZeroSourceNode) notifyAddedAsChildOf(parent GraphNode) {
}

func NewZeroSourceNode() *ZeroSourceNode {
	return &ZeroSourceNode{
		state: NodeState_Ready,
		stop:  make(chan struct{}),
	}
}

// Writes all input to stdout
type StdoutSinkNode struct {
	in GraphNode
}

func (stdoutSink *StdoutSinkNode) Read(p []byte) (int, error) {
	panic("cannot read from StdoutSinkNode")
}

func (stdoutSink *StdoutSinkNode) Write(p []byte) (int, error) {
	return fmt.Println("%v", p)
}

func (stdoutSink *StdoutSinkNode) Start() error {
	if stdoutSink.in == nil {
		panic("cannot start StdoutSinkNode because 'in' node is nil")
	}

	err := stdoutSink.in.Start()
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 0xffff)

		for stdoutSink.State() == NodeState_Running {
			n, err := stdoutSink.in.Read(buf)
			if err != nil {
				log.Println("failed to read from StdoutSinkNode", err)
				stdoutSink.Stop()
				break
			}

			buf = buf[:n]

			_, err = stdoutSink.Write(buf)
			if err != nil {
				log.Println("failed to write to StdoutSinkNode", err)
				stdoutSink.Stop()
				break
			}
		}
	}()

	return nil
}

func (stdoutSink *StdoutSinkNode) Stop() error {
	if stdoutSink.in == nil {
		panic("cannot stop StdoutSinkNode because 'in' node is nil")
	}

	return stdoutSink.in.Stop()
}

func (stdoutSink *StdoutSinkNode) Wait() {
	if stdoutSink.in == nil {
		panic("cannot wait StdoutSinkNode because 'in' node is nil")
	}

	stdoutSink.in.Wait()
}

func (stdoutSink *StdoutSinkNode) State() NodeState {
	if stdoutSink.in == nil {
		panic("cannot get state of StdoutSinkNode because 'in' node is nil")
	}

	return stdoutSink.in.State()
}

func (stdoutSink *StdoutSinkNode) GetParentNodes() []GraphNode {
	if stdoutSink.in == nil {
		return []GraphNode{}
	}

	return []GraphNode{stdoutSink.in}
}

func (stdoutSink *StdoutSinkNode) GetChildNodes() []GraphNode {
	return []GraphNode{}
}

func (stdoutSink *StdoutSinkNode) AddParent(parent GraphNode) {
	if stdoutSink.in != nil {
		panic("StdoutSinkNode cannot have more than 1 parent")
	}

	stdoutSink.in = parent
	parent.notifyAddedAsParentOf(stdoutSink)
}

func (stdoutSink *StdoutSinkNode) AddChild(child GraphNode) {
	panic("cannot add child to StdoutSinkNode")
}

func (stdoutSink *StdoutSinkNode) notifyAddedAsParentOf(child GraphNode) {
}

func (stdoutSink *StdoutSinkNode) notifyAddedAsChildOf(parent GraphNode) {
	stdoutSink.in = parent
}

func NewStdoutSinkNode() *StdoutSinkNode {
	return &StdoutSinkNode{}
}
