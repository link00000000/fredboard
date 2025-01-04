package graph

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"accidentallycoded.com/fredboard/v3/audio/codecs"
)

type NodeState byte

const (
	NodeState_NotReady NodeState = iota
	NodeState_Running
	NodeState_Stopped
)

type GraphNode interface {
	io.Reader

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
	return os.Stdout.Write(p)
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
			if err == io.EOF {
				continue
			}

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

// Reads a file from the file system
type FSFileSourceNode struct {
	out GraphNode
	fd  *os.File

	cancelCtx context.CancelFunc
	ctx       context.Context
}

func (fsFileSource *FSFileSourceNode) Open(name string, ctx context.Context) error {
	fsFileSource.ctx, fsFileSource.cancelCtx = context.WithCancel(ctx)

	fd, err := os.Open(name)
	if err != nil {
		return err
	}

	fsFileSource.fd = fd

	return nil
}

func (fsFileSource *FSFileSourceNode) Close() error {
	return fsFileSource.fd.Close()
}

func (fsFileSource *FSFileSourceNode) Wait() error {
	<-fsFileSource.ctx.Done()
	return fsFileSource.ctx.Err()
}

func (fsFileSource *FSFileSourceNode) Read(p []byte) (int, error) {
	n, err := fsFileSource.fd.Read(p)

	if err == io.EOF {
		fsFileSource.cancelCtx()
	}

	return n, err
}

func (fsFileSource *FSFileSourceNode) Write(p []byte) (int, error) {
	panic("cannot write to FSFileSourceNode")
}

func (fsFileSource *FSFileSourceNode) GetParentNodes() []GraphNode {
	return []GraphNode{}
}

func (fsFileSource *FSFileSourceNode) GetChildNodes() []GraphNode {
	if fsFileSource.out == nil {
		return []GraphNode{}
	}

	return []GraphNode{fsFileSource.out}
}

func (fsFileSource *FSFileSourceNode) AddParent(parent GraphNode) {
	panic("cannot add parent to FSFileSourceNode")
}

func (fsFileSource *FSFileSourceNode) AddChild(child GraphNode) {
	if fsFileSource.out != nil {
		panic("FSFileSource cannot have more than 1 child")
	}

	fsFileSource.out = child
	child.notifyAddedAsChildOf(fsFileSource)
}

func (fsFileSource *FSFileSourceNode) notifyAddedAsParentOf(child GraphNode) {
	fsFileSource.out = child
}

func (fsFileSource *FSFileSourceNode) notifyAddedAsChildOf(parent GraphNode) {
}

func NewFSFileSource() *FSFileSourceNode {
	return &FSFileSourceNode{}
}

// Writes a file to the file system
type FSFileSinkNode struct {
	in GraphNode
	fd *os.File

	onInNodeChanged chan struct{}

	cancelCtx context.CancelFunc
	ctx       context.Context
}

func (fsFileSink *FSFileSinkNode) Open(name string, ctx context.Context) error {
	fsFileSink.ctx, fsFileSink.cancelCtx = context.WithCancel(ctx)

	fd, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0)
	if err != nil {
		return err
	}

	fsFileSink.fd = fd

	return nil
}

func (fsFileSink *FSFileSinkNode) Close() error {
	return fsFileSink.fd.Close()
}

func (fsFileSink *FSFileSinkNode) Wait() error {
	<-fsFileSink.ctx.Done()
	return fsFileSink.ctx.Err()
}

func (fsFileSink *FSFileSinkNode) Read(p []byte) (int, error) {
	panic("cannot read from FSFileSinkNode")
}

func (fsFileSink *FSFileSinkNode) Write(p []byte) (int, error) {
	n, err := fsFileSink.fd.Write(p)

	if err == io.EOF {
		fsFileSink.cancelCtx()
	}

	return n, err
}

func (fsFileSink *FSFileSinkNode) Start(ctx context.Context) error {
	go func() {
		buf := make([]byte, 0xffff)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if fsFileSink.in == nil {
				select {
				case <-fsFileSink.onInNodeChanged:
					continue
				case <-ctx.Done():
					return
				}
			}

			n, err := fsFileSink.in.Read(buf)
			if err == io.EOF {
				fsFileSink.cancelCtx()
				return
			}

			if err != nil {
				// TODO: Do something with error
				log.Println("failed to read from StdoutSinkNode", err)
				return
			}

			buf = buf[:n]

			_, err = fsFileSink.Write(buf)
			if err != nil {
				// TODO: Do something with error
				log.Println("failed to write to StdoutSinkNode", err)
				return
			}
		}
	}()

	return nil
}

func (fsFileSink *FSFileSinkNode) GetParentNodes() []GraphNode {
	return []GraphNode{}
}

func (fsFileSink *FSFileSinkNode) GetChildNodes() []GraphNode {
	if fsFileSink.in == nil {
		return []GraphNode{}
	}

	return []GraphNode{fsFileSink.in}
}

func (fsFileSink *FSFileSinkNode) AddParent(parent GraphNode) {
	if fsFileSink.in != nil {
		panic("FSFileSink cannot have more than 1 parent")
	}

	fsFileSink.in = parent
	parent.notifyAddedAsParentOf(fsFileSink)

	select {
	case fsFileSink.onInNodeChanged <- struct{}{}:
	default:
	}
}

func (fsFileSink *FSFileSinkNode) AddChild(child GraphNode) {
	panic("cannot add child to FSFileSinkNode")
}

func (fsFileSink *FSFileSinkNode) notifyAddedAsParentOf(child GraphNode) {
}

func (fsFileSink *FSFileSinkNode) notifyAddedAsChildOf(parent GraphNode) {
	fsFileSink.in = parent
	select {
	case fsFileSink.onInNodeChanged <- struct{}{}:
	default:
	}
}

func NewFSFileSink() *FSFileSinkNode {
	return &FSFileSinkNode{}
}

// Transcodes signed 16-bit PCM raw audio to raw Opus
type PCMS16LE_Opus_TranscoderNode struct {
	in  GraphNode
	out GraphNode

	encoder *codecs.OpusEncoder
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) Read(p []byte) (int, error) {
	if transcoder.in == nil {
		panic("cannot read from PCMS16LE_Opus_TranscoderNode that does not have an 'in' node")
	}

	buf := bytes.NewBuffer(p)
	buf.Reset()

	n, err := transcoder.encoder.EncodeFromPCMS16LE(transcoder.in, buf, 960)
	return n, err
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) Initialize() error {
	encoder, err := codecs.NewOpusEncoder(48000, 2)
	if err != nil {
		return err
	}

	transcoder.encoder = encoder

	return nil
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) GetParentNodes() []GraphNode {
	if transcoder.in == nil {
		return []GraphNode{}
	}

	return []GraphNode{transcoder.in}
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) GetChildNodes() []GraphNode {
	if transcoder.out == nil {
		return []GraphNode{}
	}

	return []GraphNode{transcoder.out}
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) AddParent(parent GraphNode) {
	if transcoder.in != nil {
		panic("PCMS16LE_Opus_TranscoderNode cannot have more than 1 parent")
	}

	transcoder.in = parent
	parent.notifyAddedAsParentOf(transcoder)
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) AddChild(child GraphNode) {
	if transcoder.out != nil {
		panic("PCMS16LE_Opus_TranscoderNode cannot have more than 1 child")
	}

	transcoder.out = child
	child.notifyAddedAsChildOf(transcoder)
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) notifyAddedAsParentOf(child GraphNode) {
	transcoder.out = child
}

func (transcoder *PCMS16LE_Opus_TranscoderNode) notifyAddedAsChildOf(parent GraphNode) {
	transcoder.in = parent
}

func NewPCMS16LE_Opus_TranscoderNode() *PCMS16LE_Opus_TranscoderNode {
	return &PCMS16LE_Opus_TranscoderNode{}
}
