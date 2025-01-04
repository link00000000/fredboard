package graph

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"accidentallycoded.com/fredboard/v3/gaps/xmath"
	"layeh.com/gopus"
)

var (
	ErrFileNotOpen = errors.New("file not open")
)

var (
	_ AudioGraphNode = (*FSFileSourceNode)(nil)
	_ AudioGraphNode = (*FSFileSinkNode)(nil)
	_ AudioGraphNode = (*GainNode)(nil)
	_ AudioGraphNode = (*MixerNode)(nil)
	_ AudioGraphNode = (*PCMS16LE_Opus_TranscoderNode)(nil)
)

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

func (node *FSFileSourceNode) PreTick() error {
	return nil
}

func (node *FSFileSourceNode) PostTick() error {
	return nil
}

func (node *FSFileSourceNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNPins(ins, "in", 0, 0); err != nil {
		return fmt.Errorf("FSFileSourceNode.Tick error: %w", err)
	}

	if err := AssertNPins(outs, "out", 1, 1); err != nil {
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

func (node *FSFileSinkNode) PreTick() error {
	return nil
}

func (node *FSFileSinkNode) PostTick() error {
	return nil
}

func (node *FSFileSinkNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNPins(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("FSFileSinkNode.Tick error: %w", err)
	}

	if err := AssertNPins(outs, "out", 0, 0); err != nil {
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

func (node *GainNode) PreTick() error {
	return nil
}

func (node *GainNode) PostTick() error {
	return nil
}

func (node *GainNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNPins(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("GainNode.Tick error: %w", err)
	}

	if err := AssertNPins(outs, "out", 1, 1); err != nil {
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

func (node *MixerNode) PreTick() error {
	return nil
}

func (node *MixerNode) PostTick() error {
	return nil
}

func (node *MixerNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNPins(ins, "in", 2, 2); err != nil {
		return fmt.Errorf("MixerNode.Tick error: %w", err)
	}

	if err := AssertNPins(outs, "out", 1, 1); err != nil {
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

// Opus is output using frame-length-encoding
// 1. The first two bytes are an int16 little endian representing the length of the frame
// 2. The next n bytes are the frame data
type PCMS16LE_Opus_TranscoderNode struct {
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

func (node *PCMS16LE_Opus_TranscoderNode) PreTick() error {
	return nil
}

func (node *PCMS16LE_Opus_TranscoderNode) PostTick() error {
	return nil
}

func (node *PCMS16LE_Opus_TranscoderNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := AssertNPins(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("PCMS16LE_Opus_TranscoderNode.Tick error: %w", err)
	}

	if err := AssertNPins(outs, "out", 1, 1); err != nil {
		return fmt.Errorf("PCMS16LE_Opus_TranscoderNode error: %w", err)
	}

	for {
		// Encode block to opus and write to out
		if len(node.currentPCMSampleBlock) == node.frameSize() {
			// TODO: Reuse opus buffer
			opus, err := node.opusEncoder.Encode(node.currentPCMSampleBlock, node.frameSize(), node.frameSize())
			if err != nil {
				return fmt.Errorf("PCMS16LE_Opus_TranscoderNode encoding error: %w", err)
			}

			encodedFrameLength := int16(len(opus))
			err = binary.Write(outs[0], binary.LittleEndian, encodedFrameLength)
			if err != nil {
				return fmt.Errorf("PCMS16LE_Opus_TranscoderNode.Tick write error: %w", err)
			}

			_, err = outs[0].Write(opus)
			if err != nil {
				return fmt.Errorf("PCMS16LE_Opus_TranscoderNode.Tick write error: %w", err)
			}

			node.currentPCMSampleBlock = node.currentPCMSampleBlock[:0]

			break
		}

		// Read the next two bytes and convert to int16 sample and add to block
		if node.currentPCMSampleBufOffset == len(node.currentPCMSampleBuf) {
			sample := int16((node.currentPCMSampleBuf[0] << 8) | node.currentPCMSampleBuf[1])
			node.currentPCMSampleBlock = append(node.currentPCMSampleBlock, sample)

			node.currentPCMSampleBufOffset = 0
		}

		n, err := ins[0].Read(node.currentPCMSampleBuf[:node.currentPCMSampleBufOffset])
		node.currentPCMSampleBufOffset += n

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return fmt.Errorf("PCMS16LE_Opus_TranscoderNode.Tick read error: %w", err)
		}
	}

	return nil
}

// frame size is measured in samples
func (node *PCMS16LE_Opus_TranscoderNode) frameSize() int {
	return int(node.frameDuration.Seconds() * float64(node.sampleRate))
}

func NewPCM16LE_Opus_TransoderNode(sampleRate, nChannels int, frameDuration time.Duration) *PCMS16LE_Opus_TranscoderNode {
	opusEncoder, err := gopus.NewEncoder(sampleRate, nChannels, gopus.Audio)
	if err != nil {
		panic(fmt.Sprintf("failed to create opus encoder with gopus.NewEncoder: %w", err))
	}

	node := &PCMS16LE_Opus_TranscoderNode{
		sampleRate:                sampleRate,
		nChannels:                 nChannels,
		frameDuration:             frameDuration,
		opusEncoder:               opusEncoder,
		currentPCMSampleBufOffset: 0,
	}

	node.currentPCMSampleBlock = make([]int16, 0, node.frameSize()*nChannels)

	return node
}
