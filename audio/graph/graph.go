package graph

import (
	"context"
	"fmt"
	"io"
	"log"
)

type NodeState byte

const (
	NodeState_NotReady NodeState = iota
	NodeState_Running
	NodeState_Stopped
)

type GraphNode interface {
	io.ReadWriter

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

// Reads zeroes until stopped
type ZeroSourceNode struct {
	out GraphNode
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
	return &ZeroSourceNode{}
}

// Writes all input to stdout
type StdoutSinkNode struct {
	in              GraphNode
	onInNodeChanged chan struct{}
}

func (stdoutSink *StdoutSinkNode) Read(p []byte) (int, error) {
	panic("cannot read from StdoutSinkNode")
}

func (stdoutSink *StdoutSinkNode) Write(p []byte) (int, error) {
	return fmt.Println("%v", p)
}

func (stdoutSink *StdoutSinkNode) Start(ctx context.Context) error {
	go func() {
		buf := make([]byte, 0xffff)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if stdoutSink.in == nil {
				select {
				case <-stdoutSink.onInNodeChanged:
					continue
				case <-ctx.Done():
					return
				}
			}

			n, err := stdoutSink.in.Read(buf)
			if err != nil {
				// TODO: Do something with error
				log.Println("failed to read from StdoutSinkNode", err)
				return
			}

			buf = buf[:n]

			_, err = stdoutSink.Write(buf)
			if err != nil {
				// TODO: Do something with error
				log.Println("failed to write to StdoutSinkNode", err)
				return
			}
		}
	}()

	return nil
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

	select {
	case stdoutSink.onInNodeChanged <- struct{}{}:
	default:
	}
}

func (stdoutSink *StdoutSinkNode) AddChild(child GraphNode) {
	panic("cannot add child to StdoutSinkNode")
}

func (stdoutSink *StdoutSinkNode) notifyAddedAsParentOf(child GraphNode) {
}

func (stdoutSink *StdoutSinkNode) notifyAddedAsChildOf(parent GraphNode) {
	stdoutSink.in = parent
	select {
	case stdoutSink.onInNodeChanged <- struct{}{}:
	default:
	}
}

func NewStdoutSinkNode() *StdoutSinkNode {
	return &StdoutSinkNode{
		onInNodeChanged: make(chan struct{}),
	}
}
