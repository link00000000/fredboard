package sources

import (
	"io"
	"os"

	"accidentallycoded.com/fredboard/v3/codecs"
)

type FileStream struct {
  fd *os.File
  reader *codecs.OggOpusReader
  done chan bool
}

func NewFileStream(path string) (*FileStream, error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }

  return &FileStream{ fd: f }, nil
}


// Implements [io.Closer]
func (s *FileStream) Close() error {
  return s.fd.Close()
}

func (s *FileStream) Start(dataChannel chan[]byte, errChannel chan error) error {
  s.reader = codecs.NewOggOpusReader(s.fd)
  s.done = make(chan bool, 1)

  go func() {
    for {
      _, pkt, err := s.reader.ReadNextOpusPacket()

      switch {
      case err == io.EOF:
        s.done <- true
        return
      case err != nil:
        errChannel <- err
        return
      }

      dataChannel <- pkt
    }
  }()

  return nil
}

func (s *FileStream) Wait() error {
  <- s.done

  return nil
}
