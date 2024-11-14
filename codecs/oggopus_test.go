package codecs

import (
	"io"
	"os"
	"testing"

  "accidentallycoded.com/fredboard/v3/codecs/test_samples"
)

func TestRead(t *testing.T) {
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

  r := NewOggOpusReader(f)
  pktIdx := 0
  segIdx := 0

  for {
    _, pkt, err := r.ReadNextPacket()
    if err == io.EOF {
      break
    }

    if err != nil {
      t.Fatal("Failed to read next packet", err)
    }

    for i, segment := range pkt.Segments {
      sampleSegment := test_samples.TestSample[segIdx]

      if len(segment) != len(sampleSegment) {
        t.Fatalf("Segment is not the correct length. Expected %d, got %d", len(sampleSegment), len(segment))
      }

      for j, b := range segment {
        if b != sampleSegment[i] {
          t.Fatalf("Incorrect value at packet %d, segment %d, byte %d (segment %d, byte %d in sample). Expected %d, got %d", pktIdx, i, j, segIdx, j, sampleSegment[j], segment[j])
        }
      }

      segIdx++
    }

    pktIdx++
  }
}
