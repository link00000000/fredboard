package codecs

import (
	"encoding/binary"
	"io"

	"layeh.com/gopus"
)

type OpusEncoder struct {
	sampleRate      int
	nChannels       int
	internalEncoder *gopus.Encoder
}

func NewOpusEncoder(sampleRate, nChannels int) (*OpusEncoder, error) {
	internalEncoder, err := gopus.NewEncoder(sampleRate, nChannels, gopus.Audio)

	if err != nil {
		return nil, err
	}

	return &OpusEncoder{nChannels: nChannels, internalEncoder: internalEncoder}, nil
}

func (e *OpusEncoder) EncodePCMS16LE(reader io.Reader, writer io.Writer, frameSize int) error {
	for {
		pcmBuf := make([]int16, e.nChannels*frameSize)
		err := binary.Read(reader, binary.LittleEndian, &pcmBuf)

		if err == io.EOF {
			break
		}

		if err != nil && err != io.ErrUnexpectedEOF {
			return err
		}

		eof := err == io.ErrUnexpectedEOF
		opusBuf, err := e.internalEncoder.Encode(pcmBuf, frameSize, frameSize*e.nChannels*2)

		if err != nil {
			return err
		}

		_, err = writer.Write(opusBuf)

		if err != nil {
			return err
		}

		if eof {
			break
		}
	}

	return nil
}

func (e *OpusEncoder) EncodeDCA0(reader io.Reader, writer io.Writer) error {
  for {
    var segmentLength uint16
    err := binary.Read(reader, binary.LittleEndian, &segmentLength)

    if err == io.EOF {
      break
    }

    if err != nil {
      return err
    }

    data := make([]byte, segmentLength)
    _, err = reader.Read(data)

    if err != nil {
      return err
    }

    _, err = writer.Write(data)
    
    if err != nil {
      return err
    }
  }

  return nil
}
