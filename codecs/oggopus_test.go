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

  for {
    _, pkt, err := r.ReadNextOpusPacket()
    if err == io.EOF {
      if pktIdx < len(testdata.TestSample) {
        t.Fatalf("Expected more data. Stopped at packet %d, should be %d packets", pktIdx, len(testdata.TestSample))
      }

      break
    }

    if err != nil {
      t.Fatal("Failed to read next packet", err)
    }

    samplePkt := testdata.TestSample[pktIdx]

    if len(pkt) != len(samplePkt) {
      t.Fatalf("Packet %d is not the correct length. Expected %d, got %d", pktIdx, len(samplePkt), len(pkt))
    }

    for i, b := range pkt {
      if b != samplePkt[i] {
        t.Fatalf("Incorrect value at packet %d, byte %d. Expected %d, got %d", pktIdx, i, samplePkt[i], b)
      }
    }

    pktIdx++
  }
}
