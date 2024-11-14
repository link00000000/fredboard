package codecs

import (
	"io"
	"os"
	"testing"

  "accidentallycoded.com/fredboard/v3/codecs/test_samples"
)

func TestDCA0Reader(t *testing.T) {
  f, err := os.Open("./test_samples/sample.dca")

  if err != nil {
    t.Fatal(err)
  }

  defer func() {
    err := f.Close()
    if err != nil {
      t.Fatal(err)
    }
  }()

  r := NewDCA0Reader(f)
  segIdx := 0

  for {
    _, segment, err := r.ReadNextSegment()
    if err == io.EOF {
      break
    }

    if err != nil {
      t.Fatal("Failed to read next segment", err)
    }

    sampleSegment := test_samples.TestSample[segIdx]

    if len(segment) != len(sampleSegment) {
      t.Fatalf("Segment is not the correct length. Expected %d, got %d", len(sampleSegment), len(segment))
    }

    for i, b := range segment {
      if b != sampleSegment[i] {
        t.Fatalf("Incorrect value at segment %d, byte %d. Expected %d, got %d", segIdx, i, sampleSegment[i], b)
      }
    }

    segIdx++
  }
}
