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
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

var (
	_ AudioGraphNode = (*FSFileSourceNode)(nil)
	_ AudioGraphNode = (*FSFileSinkNode)(nil)
	_ AudioGraphNode = (*GainNode)(nil)
	_ AudioGraphNode = (*MixerNode)(nil)
	_ AudioGraphNode = (*OpusEncoderNode)(nil)
	_ AudioGraphNode = (*CompositeNode)(nil)
	_ AudioGraphNode = (*ZeroSourceNode)(nil)
	_ AudioGraphNode = (*DiscordSinkNode)(nil)
)

var (
	ErrFileNotOpen   = errors.New("file not open")
	ErrNotExist      = errors.New("does not exist")
	ErrAlreadyExists = errors.New("already exists")
)

type NodeIOBound int

const (
	NodeIOBound_Unbounded = -1
)

func (bound NodeIOBound) String() string {
	if bound == NodeIOBound_Unbounded {
		return "unbounded"
	}

	return fmt.Sprintf("%d", int(bound))
}

type NodeIOType byte

const (
	NodeIOType_In = iota
	NodeIOType_Out
)

func (ioType NodeIOType) String() string {
	switch ioType {
	case NodeIOType_In:
		return "In"
	case NodeIOType_Out:
		return "Out"
	}

	panic(fmt.Sprintf("invalid NodeIOType 0x%x", ioType))
}

type NodeIOBoundError struct {
	ioType  NodeIOType
	nMin    NodeIOBound
	nMax    NodeIOBound
	nActual int
}

func AssertNodeIOBounds[T any](ios []T, ioType NodeIOType, nMin, nMax NodeIOBound) *NodeIOBoundError {
	nActual := len(ios)

	if nMin != NodeIOBound_Unbounded && nActual < int(nMin) {
		return NewNodeIOBoundError(ioType, nMin, nMax, nActual)
	}

	if nMax != NodeIOBound_Unbounded && nActual > int(nMax) {
		return NewNodeIOBoundError(ioType, nMin, nMax, nActual)
	}

	return nil
}

func (err NodeIOBoundError) Error() string {
	return fmt.Sprintf("invalid IO configuration: type = %s, min = %s, max = %s, actual = %d", err.ioType.String(), err.nMin.String(), err.nMax.String(), err.nActual)
}

func NewNodeIOBoundError(ioType NodeIOType, nMin, nMax NodeIOBound, nActual int) *NodeIOBoundError {
	return &NodeIOBoundError{ioType, nMin, nMax, nActual}
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
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 0, 0); err != nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
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
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, 1); err != nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 0, 0); err != nil {
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
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, 1); err != nil {
		return fmt.Errorf("GainNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
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

// 2 or more inputs, mixes all inputs together into a single output
// If there is only 1 input, it is passed directly through to the output
type MixerNode struct {
}

func (node *MixerNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, NodeIOBound_Unbounded); err != nil {
		return fmt.Errorf("MixerNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
		return fmt.Errorf("MixerNode.Tick error: %w", err)
	}

	eofs := make([]bool, len(ins))
	samples := make([]int16, len(ins))
	for {
		allInsEof := true
		for _, eof := range eofs {
			if !eof {
				allInsEof = false
			}
		}

		if allInsEof {
			break
		}

		for i := range ins {
			if eofs[i] {
				samples[i] = 0
				continue
			}

			err := binary.Read(ins[i], binary.LittleEndian, &samples[i])

			if err == io.EOF {
				eofs[i] = true
				samples[i] = 0
				continue
			}

			if err != nil {
				return fmt.Errorf("MixerNode.Tick read error: %w", err)
			}
		}

		mixedSample := mixSamples(samples)
		err := binary.Write(outs[0], binary.LittleEndian, mixedSample)
		if err != nil {
			return fmt.Errorf("MixerNode.Tick write error: %w", err)
		}
	}

	return nil
}

// Possibly update to prevent clipping. See https://stackoverflow.com/a/27000317
// I don't know if that algorithm will work
func mixSamples(samples []int16) int16 {
	var mixedSample int16 = 0

	for _, sample := range samples {
		var temp int32 = int32(sample) + int32(mixedSample)
		mixedSample = int16(xmath.Clamp(temp, int32(math.MinInt16), int32(math.MaxInt16)))
	}

	return mixedSample
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
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, 1); err != nil {
		return fmt.Errorf("OpusEncoderNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
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
		if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, 1); err != nil {
			return fmt.Errorf("CompositeNode.Tick error: %w", err)
		}
	}

	if node.outNode != nil {
		if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
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
				return fmt.Errorf("CompositeNode.tickInternalNode error: %w", err)
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

type ZeroSourceNode struct {
	desiredBufLen uint64
}

func (node *ZeroSourceNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 0, 0); err != nil {
		return fmt.Errorf("ZeroSourceNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 1, 1); err != nil {
		return fmt.Errorf("ZeroSourceNode.Tick error: %w", err)
	}

	zero := [1]byte{0x00}
	for range node.desiredBufLen {
		outs[0].Write(zero[:])
	}

	return nil
}

func NewZeroSourceNode(desiredBufLen uint64) *ZeroSourceNode {
	return &ZeroSourceNode{desiredBufLen: desiredBufLen}
}

type DiscordSinkNode struct {
	conn *discordgo.VoiceConnection
}

func (node *DiscordSinkNode) PreTick() error {
	return nil
}

func (node *DiscordSinkNode) PostTick() error {
	return nil
}

func (node *DiscordSinkNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNodeIOBounds(ins, NodeIOType_In, 1, 1); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	if err := AssertNodeIOBounds(outs, NodeIOType_Out, 0, 0); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	for {
		var encodedFrameSize int16
		err := binary.Read(ins[0], binary.LittleEndian, &encodedFrameSize)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return fmt.Errorf("DiscordSinkNode.Tick read error: %w", err)
		}

		// TODO: Cache buffer usedf or p?
		p, err := io.ReadAll(io.LimitReader(ins[0], int64(encodedFrameSize)))
		if err != nil {
			return fmt.Errorf("DiscordSinkNode.Tick read error: %w", err)
		}

		node.conn.OpusSend <- p
	}

	return nil
}

func NewDiscordSinkNode(conn *discordgo.VoiceConnection) *DiscordSinkNode {
	return &DiscordSinkNode{conn: conn}
}
