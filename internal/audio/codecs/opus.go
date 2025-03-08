package codecs

import (
	"fmt"
	"io"
	"time"

	"layeh.com/gopus"
)

type OpusWriter interface {
	Write(p [][]byte) (n int /* number of frames consumed */, err error)
}

type opusEncoderWriter struct {
	w   OpusWriter
	enc *gopus.Encoder

	nChannels    int
	sampleRateHz int
	frameSize    int

	pcm []byte
}

func (e *opusEncoderWriter) Write(p []byte) (n int, err error) {
	e.pcm = append(e.pcm, p...)
	frames := make([][]byte, 0)

	frameSizeBytes := e.nChannels * e.frameSize * 2 // last *2 is because each sample is an int16 and thus 2 bytes
	for len(e.pcm) >= frameSizeBytes {
		pcm := BytesToS16LE(e.pcm[:frameSizeBytes])
		opus, err := e.enc.Encode(pcm, e.frameSize, frameSizeBytes)
		e.pcm = e.pcm[frameSizeBytes:]

		if err != nil {
			return len(p), err
		}

		frames = append(frames, opus)
	}

	for len(frames) > 0 {
		n, err := e.w.Write(frames)
		frames = frames[n:]

		if err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

// flush remaining pcm data in into a final padded opus frame
func (e *opusEncoderWriter) Close() error {
	pcm := BytesToS16LE(e.pcm)
	frameSizeMs := float32(len(pcm)) / float32(e.nChannels) * 1000 / float32(e.sampleRateHz)

	var targetMs float32
	for _, validMs := range []float32{2.5, 5, 10, 20, 40, 60} {
		if frameSizeMs <= validMs {
			targetMs = validMs
			break
		}
	}

	targetFrameSize := int(targetMs * float32(e.sampleRateHz) / 1000 * float32(e.nChannels))
	paddedPCM := make([]int16, targetFrameSize)
	copy(paddedPCM, pcm)

	frameSizeBytes := e.nChannels * e.frameSize * 2 // last *2 is because each sample is an int16 and thus 2 bytes
	opus, err := e.enc.Encode(pcm, e.frameSize, frameSizeBytes)
	e.pcm = e.pcm[:0]

	if err != nil {
		return err
	}

	_, err = e.w.Write([][]byte{opus})
	if err != nil {
		return err
	}

	return nil
}

// encodes opus from 16-bit signed little endian PCM
func NewOpusEncoderWriter(w OpusWriter, nChannels, sampleRateHz, frameSize int) (io.WriteCloser, error) {
	enc, err := gopus.NewEncoder(sampleRateHz, nChannels, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}

	return &opusEncoderWriter{
		w:            w,
		enc:          enc,
		nChannels:    nChannels,
		sampleRateHz: sampleRateHz,
		frameSize:    frameSize,
		pcm:          make([]byte, 0, nChannels*frameSize),
	}, nil
}

func frameDuration(frameSize, sampleRateHz int) time.Duration {
	return time.Duration((float64(frameSize) / float64(sampleRateHz)) * float64(time.Second))
}
