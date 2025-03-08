package codecs_test

import (
	"io"
	"os"
	"slices"
	"testing"

	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/audio/codecs/testdata"
)

type opusBuffer struct {
	Frames [][]byte
}

func (b *opusBuffer) Write(p [][]byte) (n int /* number of frames consumed */, err error) {
	if b.Frames == nil {
		b.Frames = make([][]byte, 0)
	}

	b.Frames = append(b.Frames, p...)
	return len(p), nil
}

func TestOpusEncoderWriter(t *testing.T) {
	f, err := os.Open("./testdata/sample.pcms16le")
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	var buffer opusBuffer
	w, err := codecs.NewOpusEncoderWriter(&buffer, 2, 48000, 960)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(w, f)
	if err != nil {
		t.Fatal(err)
	}

	w.Close()

	if len(buffer.Frames) != len(testdata.PCMS16LESampleEncodedAsOpus) {
		t.Fatalf("incorrect number of frames. want %d, got %d", len(testdata.PCMS16LESampleEncodedAsOpus), len(buffer.Frames))
	}

	for idx, frame := range buffer.Frames {
		if !slices.Equal(frame, testdata.PCMS16LESampleEncodedAsOpus[idx]) {
			t.Fatalf("incorrect frame (idx = %d). want %v, got %v", idx, testdata.PCMS16LESampleEncodedAsOpus[idx], frame)
		}
	}
}
