package graph

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"time"

	xmath "accidentallycoded.com/fredboard/v3/internal/math"
	"layeh.com/gopus"
)

var (
	_ AudioGraphNode = (*FSFileSourceNode)(nil)
	_ AudioGraphNode = (*FSFileSinkNode)(nil)
	_ AudioGraphNode = (*GainNode)(nil)
	_ AudioGraphNode = (*MixerNode)(nil)
	_ AudioGraphNode = (*OpusEncoderNode)(nil)
	_ AudioGraphNode = (*CompositeNode)(nil)
)

var (
	ErrFileNotOpen   = errors.New("file not open")
	ErrNotExist      = errors.New("does not exist")
	ErrAlreadyExists = errors.New("already exists")
)

type InvalidNNodeIO struct {
	ioType  string
	nMin    int
	nMax    int
	nActual int
}

func AssertNNodeIO[T any](ios []T, ioType string, nMin, nMax int) *InvalidNNodeIO {
	nActual := len(ios)

	if nActual < nMin || nActual > nMax {
		return NewInvalidNNodeIOError(ioType, nMin, nMax, nActual)
	}

	return nil
}

func (err InvalidNNodeIO) Error() string {
	return fmt.Sprintf("invalid number of %s IOs: min = %d, max = %d, actual = %d", err.ioType, err.nMin, err.nMax, err.nActual)
}

func NewInvalidNNodeIOError(ioType string, nMin, nMax, nActual int) *InvalidNNodeIO {
	return &InvalidNNodeIO{ioType, nMin, nMax, nActual}
}

type AudioGraphNode interface {
	Tick(ins []io.Reader, outs []io.Writer) error
}

type FSFileSourceNode struct {
	fd    *os.File
	isEOF bool

	OnEOF func()
}

func (node *FSFileSourceNode) OpenFile(name string) error {
	fd, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("FSFileSourceNode.OpenFile error: %w", err)
	}

	node.fd = fd

	return nil
}

func (node *FSFileSourceNode) CloseFile() error {
	if node.fd == nil {
		return fmt.Errorf("FSFileSourceNode.CloseFile error: %w", ErrFileNotOpen)
	}

	err := node.fd.Close()
	if err != nil {
		return fmt.Errorf("FSFileSourceNode.CloseFile error: %w", err)
	}

	return nil
}

func (node *FSFileSourceNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNNodeIO(ins, "in", 0, 0); err != nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", err)
	}

	if err := AssertNNodeIO(outs, "out", 1, 1); err != nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", err)
	}

	if node.isEOF {
		return nil
	}

	if node.fd == nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", ErrFileNotOpen)
	}

	_, err := io.CopyN(outs[0], node.fd, 1024)

	if err == io.EOF {
		node.isEOF = true
		if node.OnEOF != nil {
			node.OnEOF()
		}

		return nil
	}

	if err == io.ErrUnexpectedEOF {
		node.isEOF = true
		if node.OnEOF != nil {
			node.OnEOF()
		}
	}

	if err != nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", err)
	}

	return nil
}

func NewFSFileSourceNode() *FSFileSourceNode {
	return &FSFileSourceNode{isEOF: false}
}

type FSFileSinkNode struct {
	fd *os.File
}

func (node *FSFileSinkNode) OpenFile(name string) error {
	fd, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("FSFileSinkNode.OpenFile error: %w", err)
	}

	node.fd = fd

	return nil
}

func (node *FSFileSinkNode) CloseFile() error {
	if node.fd == nil {
		return fmt.Errorf("FSFileSinkNode.CloseFile error: %w", ErrFileNotOpen)
	}

	err := node.fd.Close()
	if err != nil {
		return fmt.Errorf("FSFileSinkNode.CloseFile error: %w", err)
	}

	return nil
}

func (node *FSFileSinkNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNNodeIO(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", err)
	}

	if err := AssertNNodeIO(outs, "out", 0, 0); err != nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", err)
	}

	if node.fd == nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", ErrFileNotOpen)
	}

	_, err := io.Copy(node.fd, ins[0])

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", err)
	}

	return nil
}

func NewFSFileSinkNode() *FSFileSinkNode {
	return &FSFileSinkNode{}
}

type GainNode struct {
	GainFactor float32
}

func (node *GainNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNNodeIO(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("GainNode.Tick error: %w", err)
	}

	if err := AssertNNodeIO(outs, "out", 1, 1); err != nil {
		return fmt.Errorf("GainNode.Tick error: %w", err)
	}

	for {
		var sample int16
		err := binary.Read(ins[0], binary.LittleEndian, &sample)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("GainNode.Tick read error: %w", err)
		}

		sample = int16(float32(sample) * node.GainFactor)

		err = binary.Write(outs[0], binary.LittleEndian, &sample)
		if err != nil {
			return fmt.Errorf("GainNode.Tick write error: %w", err)
		}
	}

	return nil
}

func NewGainNode(factor float32) *GainNode {
	return &GainNode{GainFactor: factor}
}

type MixerNode struct {
}

func (node *MixerNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNNodeIO(ins, "in", 2, 2); err != nil {
		return fmt.Errorf("MixerNode.Tick error: %w", err)
	}

	if err := AssertNNodeIO(outs, "out", 1, 1); err != nil {
		return fmt.Errorf("MixerNode.Tick error: %w", err)
	}

	for {
		var sampleA int16
		sourceAEOF := false
		err := binary.Read(ins[0], binary.LittleEndian, &sampleA)
		if err != nil {
			if err == io.EOF {
				sampleA = 0x0000
				sourceAEOF = true
			} else {
				return fmt.Errorf("MixerNode.Tick read error: %w", err)
			}
		}

		var sampleB int16
		sourceBEOF := false
		err = binary.Read(ins[1], binary.LittleEndian, &sampleB)
		if err != nil {
			if err == io.EOF {
				sampleB = 0x0000
				sourceBEOF = true
			} else {
				return fmt.Errorf("MixerNode.Tick read error: %w", err)
			}
		}

		if sourceAEOF && sourceBEOF {
			break
		}

		mixedSample := int16(xmath.Clamp(int32(sampleA)+int32(sampleB), int32(math.MinInt16), int32(math.MaxInt16)))
		err = binary.Write(outs[0], binary.LittleEndian, mixedSample)
		if err != nil {
			return fmt.Errorf("MixerNode.Tick write error: %w", err)
		}
	}

	return nil
}

func NewMixerNode() *MixerNode {
	return &MixerNode{}
}

// Encodes signed int16 PCM to Opus
//
// Opus is output using frame-length-encoding
// 1. The first two bytes are an int16 little endian representing the length of the frame
// 2. The next n bytes are the frame data
type OpusEncoderNode struct {
	sampleRate    int
	nChannels     int
	frameDuration time.Duration
	opusEncoder   *gopus.Encoder

	// current block of PCM samples being processed
	currentPCMSampleBlock []int16

	// current PCM sample being converted from bytes to int16
	currentPCMSampleBuf       [2]byte
	currentPCMSampleBufOffset int
}

func (node *OpusEncoderNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNNodeIO(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("OpusEncoderNode.Tick error: %w", err)
	}

	if err := AssertNNodeIO(outs, "out", 1, 1); err != nil {
		return fmt.Errorf("OpusEncoderNode error: %w", err)
	}

	for {
		// Encode block to opus and write to out
		if len(node.currentPCMSampleBlock) == node.frameSize() {
			// TODO: Reuse opus buffer
			opus, err := node.opusEncoder.Encode(node.currentPCMSampleBlock, node.frameSize(), node.frameSize())
			if err != nil {
				return fmt.Errorf("OpusEncoderNode encoding error: %w", err)
			}

			encodedFrameLength := int16(len(opus))
			err = binary.Write(outs[0], binary.LittleEndian, encodedFrameLength)
			if err != nil {
				return fmt.Errorf("OpusEncoderNode.Tick write error: %w", err)
			}

			_, err = outs[0].Write(opus)
			if err != nil {
				return fmt.Errorf("OpusEncoderNode.Tick write error: %w", err)
			}

			node.currentPCMSampleBlock = node.currentPCMSampleBlock[:0]

			break
		}

		// Read the next two bytes and convert to int16 sample and add to block
		if node.currentPCMSampleBufOffset == len(node.currentPCMSampleBuf) {
			sample := int16(binary.LittleEndian.Uint16(node.currentPCMSampleBuf[:]))
			node.currentPCMSampleBlock = append(node.currentPCMSampleBlock, sample)

			node.currentPCMSampleBufOffset = 0
		}

		n, err := ins[0].Read(node.currentPCMSampleBuf[node.currentPCMSampleBufOffset:])
		node.currentPCMSampleBufOffset += n

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return fmt.Errorf("OpusEncoderNode.Tick read error: %w", err)
		}
	}

	return nil
}

// frame size is measured in samples
func (node *OpusEncoderNode) frameSize() int {
	return int(node.frameDuration.Seconds() * float64(node.sampleRate))
}

func NewOpusEncoderNode(sampleRate, nChannels int, frameDuration time.Duration) *OpusEncoderNode {
	opusEncoder, err := gopus.NewEncoder(sampleRate, nChannels, gopus.Audio)
	if err != nil {
		panic(fmt.Sprintf("failed to create opus encoder with gopus.NewEncoder: %w", err))
	}

	node := &OpusEncoderNode{
		sampleRate:                sampleRate,
		nChannels:                 nChannels,
		frameDuration:             frameDuration,
		opusEncoder:               opusEncoder,
		currentPCMSampleBufOffset: 0,
	}

	node.currentPCMSampleBlock = make([]int16, 0, node.frameSize()*nChannels)

	return node
}

type CompositeNode struct {
	nodes       []AudioGraphNode
	connections []*AudioGraphConnection

	inNode  AudioGraphNode
	outNode AudioGraphNode
}

// TODO: Cancel with context
// TOOD: Use ins and outs
func (node *CompositeNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if node.inNode != nil {
		if err := AssertNNodeIO(ins, "in", 1, 1); err != nil {
			return fmt.Errorf("CompositeNode.Tick error: %w", err)
		}
	}

	if node.outNode != nil {
		if err := AssertNNodeIO(outs, "out", 1, 1); err != nil {
			return fmt.Errorf("CompositeNode.Tick error: %w", err)
		}
	}

	leafNodes := make([]AudioGraphNode, 0, len(node.connections))
	for _, n := range node.nodes {
		if !slices.ContainsFunc(node.connections, func(conn *AudioGraphConnection) bool { return conn.from == n }) {
			leafNodes = append(leafNodes, n)
		}
	}

	for _, n := range leafNodes {
		conns := make([]*AudioGraphConnection, 0, len(node.connections))
		for _, conn := range node.connections {
			if conn.to == n {
				conns = append(conns, conn)
			}
		}

		err := node.tickInternalNode(n, ins, outs)

		if err != nil {
			return fmt.Errorf("CompositeNode.Tick error: %w", err)
		}
	}

	return nil
}

func (node *CompositeNode) AddNode(n AudioGraphNode) error {
	if slices.Contains(node.nodes, n) {
		return fmt.Errorf("invalid node: %w", ErrAlreadyExists)
	}

	node.nodes = append(node.nodes, n)

	return nil
}

func (node *CompositeNode) SetInNode(n AudioGraphNode) error {
	if !slices.Contains(node.nodes, n) {
		return fmt.Errorf("invalid node: %w", ErrNotExist)
	}

	node.inNode = n

	return nil
}

func (node *CompositeNode) SetOutNode(n AudioGraphNode) error {
	if !slices.Contains(node.nodes, n) {
		return fmt.Errorf("invalid node: %w", ErrNotExist)
	}

	node.outNode = n

	return nil
}

// TODO: RemoveNode()

func (node *CompositeNode) CreateConnection(from, to AudioGraphNode) error {
	// TODO: Make a fast version that does not do this check
	idx := slices.IndexFunc(node.connections, func(conn *AudioGraphConnection) bool {
		return conn.from == from && conn.to == to
	})

	if idx != -1 {
		return fmt.Errorf("connection is invalid: %w", ErrAlreadyExists)
	}

	node.connections = append(node.connections, NewAudioGraphConnection(from, to))

	return nil
}

func (node *CompositeNode) RemoveConnection(from, to AudioGraphNode) error {
	del := slices.DeleteFunc(node.connections, func(conn *AudioGraphConnection) bool {
		return conn.from == from && conn.to == to
	})

	if len(del) == 0 {
		return fmt.Errorf("connection is invalid: %w", ErrNotExist)
	}

	return nil
}

func (node *CompositeNode) tickInternalNode(n AudioGraphNode, ins []io.Reader, outs []io.Writer) error {
	dependencies := make([]io.Reader, 0)
	dependents := make([]io.Writer, 0)

	// forward ins from composite node to registered in node
	if n == node.inNode {
		dependencies = append(dependencies, ins...)
	}

	// forward outs from composite node to registered out node
	if n == node.outNode {
		dependents = append(dependents, outs...)
	}

	// aggregate and process dependencies
	for _, conn := range node.connections {
		if conn.to == n {
			dependencies = append(dependencies, conn)

			err := node.tickInternalNode(conn.from, ins, outs)
			if err != nil {
				return fmt.Errorf("CompositeNode.tickInternalNode errorr: %w", err)
			}
		}
	}

	// aggregate dependents
	for _, conn := range node.connections {
		if conn.from == n {
			dependents = append(dependents, conn)
		}
	}

	// all dependencies have been ticked, now the tick node
	err := n.Tick(dependencies, dependents)
	if err != nil {
		return fmt.Errorf("CompositeNode.tickInternalNode error: %w", err)
	}

	return nil
}

func NewCompositeNode() *CompositeNode {
	return &CompositeNode{
		nodes:       make([]AudioGraphNode, 0),
		connections: make([]*AudioGraphConnection, 0),
	}
}
