package codecs

import (
	"os"
	"testing"

	"accidentallycoded.com/fredboard/v3/codecs/testdata"
)

type SegmentAggregator struct {
	Segments [][]byte
}

func NewSegmentAggregator() *SegmentAggregator {
	return &SegmentAggregator{Segments: make([][]byte, 0)}
}

func (aggregator *SegmentAggregator) Write(buf []byte) (int, error) {
	aggregator.Segments = append(aggregator.Segments, buf)
	return len(buf), nil
}

func TestEncodePCM16LE(t *testing.T) {
	inputFile, err := os.Open("./testdata/sample.pcms16le")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := inputFile.Close()

		if err != nil {
			t.Fatal(err)
		}
	}()

	encoder, err := NewOpusEncoder(48000, 2)

	if err != nil {
		t.Fatal(err)
	}

	output := NewSegmentAggregator()
	err = encoder.EncodePCMS16LE(inputFile, output, 960)

	if err != nil {
		t.Fatal(err)
	}

	if len(output.Segments) != len(testdata.PCMS16LESampleEncodedAsOpus) {
		t.Fatalf("Incorrect number of segments encoded. Expected %d segments, got %d segments", len(testdata.PCMS16LESampleEncodedAsOpus), len(output.Segments))
	}

	for i, segment := range output.Segments {
		expectedSegment := testdata.PCMS16LESampleEncodedAsOpus[i]

		if len(segment) != len(expectedSegment) {
			t.Fatalf("Output segment %d is not the correct length. Expected %d bytes, got %d bytes", i, len(expectedSegment), len(segment))
		}

		for j, b := range segment {
			if b != expectedSegment[j] {
				t.Fatalf("Output segment %d, byte %d is not the correct value. Expected 0x%02x, got 0x%02x", i, j, expectedSegment[j], b)
			}
		}
	}
}

func TestEncodeDCA0(t *testing.T) {
	inputFile, err := os.Open("./testdata/sample.dca0")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := inputFile.Close()

		if err != nil {
			t.Fatal(err)
		}
	}()

	encoder, err := NewOpusEncoder(48000, 2)

	if err != nil {
		t.Fatal(err)
	}

	output := NewSegmentAggregator()
	err = encoder.EncodeDCA0(inputFile, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(output.Segments) != len(testdata.DCA0SampleEncodedAsOpus) {
		t.Fatalf("Incorrect number of segments encoded. Expected %d segments, got %d segments", len(testdata.DCA0SampleEncodedAsOpus), len(output.Segments))
	}

	for i, segment := range output.Segments {
		expectedSegment := testdata.DCA0SampleEncodedAsOpus[i]

		if len(segment) != len(expectedSegment) {
			t.Fatalf("Output segment %d is not the correct length. Expected %d bytes, got %d bytes", i, len(expectedSegment), len(segment))
		}

		for j, b := range segment {
			if b != expectedSegment[j] {
				t.Fatalf("Output segment %d, byte %d is not the correct value. Expected 0x%02x, got 0x%02x", i, j, expectedSegment[j], b)
			}
		}
	}

}
