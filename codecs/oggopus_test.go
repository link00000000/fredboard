package codecs

import (
	"io"
	"os"
	"testing"

  "accidentallycoded.com/fredboard/v3/codecs/testdata"
)

func TestRead(t *testing.T) {
  f, err := os.Open("./testdata/sample.ogg")

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

    if segIdx == 0 || segIdx == 1 {
      // The first two packets will be metadata. Skip them.
      continue
    }

    for i, segment := range pkt.Segments {
      sampleSegment := testdata.TestSample[segIdx]

      if len(segment) != len(sampleSegment) {
        t.Fatalf("Segment at packet %d, segment %d (segment %d in sample) is not the correct length. Expected %d, got %d", pktIdx, i, segIdx, len(sampleSegment), len(segment))
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
