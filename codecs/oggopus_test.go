package codecs

import (
	"io"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
  f, err := os.Open("./test_samples/sample.ogg")
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

  for {
    n, pkt, err := r.ReadNextPacket()
    t.Logf("Read %d bytes", n)

    if err == io.EOF {
      break
    } else if err != nil {
      t.Fatal(err)
    }

    t.Logf("%#v", pkt)
  }
}
