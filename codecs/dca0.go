package codecs

import (
	"encoding/binary"
	"io"
)

type DCA0Reader struct {
  internalReader io.Reader
}

func NewDCA0Reader(reader io.Reader) *DCA0Reader {
  return &DCA0Reader{ internalReader: reader };
}

func (dr *DCA0Reader) ReadNextSegment() (int, []byte, error) {
  n := 0
  r := dr.internalReader

  bufSegLen := make([]byte, 2)
  nn, err := r.Read(bufSegLen)
  n += nn

  segLen := binary.LittleEndian.Uint16(bufSegLen)

  if err != nil {
    return n, nil, err
  }

  data := make([]byte, segLen)
  nn, err = r.Read(data)
  n += nn

  if err != nil {
    return n, nil, err
  }

  return n, data, nil
}
