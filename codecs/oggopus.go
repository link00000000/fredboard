package codecs

import (
	"io"
	"log/slog"
)

const oggCapturePatern = "OggS"

var logger = slog.Default()

type oggHeaderType uint8

const (
	oggHeaderType_continuation = 1 << iota
	oggHeaderType_beginningOfStream
	oggHeaderType_endOfStream
)

type oggPage struct {
  headerType oggHeaderType
  granulePosition uint64
  bitstreamSerialNumber uint32
  pageSequenceNumber uint32
  checksum uint32
  segments [][]byte
}

type oggReaderState uint8

const (
  oggReaderState_ready = iota
  oggReaderState_readingHeader
  oggReaderState_readingData
  oggReaderState_done
  oggReaderState_failed
)

type OggOpusReader struct {
	internalReader io.Reader
  currentPage *oggPage
  state oggReaderState
}

func NewOggOpusReader(reader io.Reader) *OggOpusReader {
	return &OggOpusReader{ internalReader: reader }
}

func (r *OggOpusReader) ReadNextOggPacket() (int, OpusPacket, error) {
  /*
  n := 0

  for {
    switch r.state {
    case oggReaderState_ready:
      r.state = oggReaderState_readingHeader

    case oggReaderState_readingHeader:
      nn, err := r.readNextOggPageHeader()
      n += nn

      if err == io.EOF {
        r.state = oggReaderState_done
        break
      }

      if err != nil {
        r.state = oggReaderState_failed
        return n, nil, err
      }

    case oggReaderState_readingData:
      nn, buf, err := r.readNextOggDataSegment()
      n += nn

      if err != nil {
        r.state = oggReaderState_failed
        return n, nil, err
      }

      return n, OpusPacket(buf), nil
    
    case oggReaderState_done:
      return 0, nil, io.EOF

    case oggReaderState_failed:
      return 0, nil, errors.New("Attempted to call ReadNextOggPacket() on failed reader")
    }
  }
}

func (oor *OggOpusReader) ReadNextPacket() (int, *oggPage, error) {
	n := 0
  r := oor.internalReader

	capturePatternBuf := make([]byte, len(oggCapturePatern)) // Capture pattern is always "OggS"
	nn, err := r.Read(capturePatternBuf)
	n += nn

	if err == nil && nn != len(oggCapturePatern) {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for capture pattern %s, got %d bytes", len(oggCapturePatern), oggCapturePatern, nn))
	}

	if err != nil {
		logger.Debug("Error while reading capture patern for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read capture pattern for packet in OggOpusReader", "capturePattern", string(capturePatternBuf))

	versionBuf := make([]byte, 1)
	nn, err = r.Read(versionBuf)
	n += nn

	if err == nil && nn != 1 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for version, got %d bytes", 1, nn))
	}

	if err != nil {
		logger.Debug("Error while reading version for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read version for packet in OggOpusReader", "version", versionBuf)

  headerTypeBuf := make([]byte, 1)
  nn, err = r.Read(headerTypeBuf)
  n += nn

  if err != nil && nn != 1 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for header type, got %d bytes", 1, nn))
  }

	if err != nil {
		logger.Debug("Error while reading header type for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read header type for packet in OggOpusReader", "headerType", headerTypeBuf)

  granulePositionBuf := make([]byte, 8)
  nn, err = r.Read(granulePositionBuf)
  n += n

  if err != nil && nn != 8 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for granule position, got %d bytes", 8, nn))
  }

	if err != nil {
		logger.Debug("Error while reading granule position for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read granule position for packet in OggOpusReader", "granulePosition", granulePositionBuf)

  bitstreamSerialNumberBuf := make([]byte, 4)
  nn, err = r.Read(bitstreamSerialNumberBuf)
  n += n

  if err != nil && nn != 4 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for bitstream serial number, got %d bytes", 4, nn))
  }

	if err != nil {
		logger.Debug("Error while reading bitstream serial number for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read bitstream serial number for packet in OggOpusReader", "bitstreamSerialNumber", bitstreamSerialNumberBuf)

  pageSequenceNumberBuf := make([]byte, 4)
  nn, err = r.Read(pageSequenceNumberBuf)
  n += n

  if err != nil && nn != 4 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for page sequence number, got %d bytes", 4, nn))
  }

	if err != nil {
		logger.Debug("Error while reading page sequence number for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read page sequence number for packet in OggOpusReader", "pageSequenceNumber", pageSequenceNumberBuf)

  checksumBuf := make([]byte, 4)
  nn, err = r.Read(checksumBuf)
  n += n

  if err != nil && nn != 4 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for checksum, got %d bytes", 4, nn))
  }

	if err != nil {
		logger.Debug("Error while reading checksum for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read checksum for packet in OggOpusReader", "checksum", checksumBuf)

  pageSegmentsBuf := make([]byte, 1)
  nn, err = r.Read(pageSegmentsBuf)
  n += n

  if err != nil && nn != 1 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for page segments, got %d bytes", 1, nn))
  }

	if err != nil {
		logger.Debug("Error while reading page segments for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read page segments for packet in OggOpusReader", "pageSegments", pageSegmentsBuf)

  segmentTable := make([]byte, int(pageSegmentsBuf[0]))
  nn, err = r.Read(segmentTable)
  n += nn

  if err != nil && nn != int(pageSegmentsBuf[0]) {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for segment table, got %d bytes", int(pageSegmentsBuf[0]), nn))
  }

  if err != nil {
		logger.Debug("Error while reading segment table for packet in OggOpusReader", "error", err)
		return n, nil, err
  }

  segmentData := make([][]byte, int(pageSegmentsBuf[0]))
  for i := 0; i < int(pageSegmentsBuf[0]); i++ {
    logger.Debug("Reading next segment", "index", i, "total", int(pageSegmentsBuf[0]))

    segmentData[i] = make([]byte, int(segmentTable[i]))
    nn, err := r.Read(segmentData[i])
    n += nn

    if err != nil && nn != int(segmentTable[i]) {
      err = errors.New(fmt.Sprintf("Expected to read %d bytes for segment data, got %d bytes", int(segmentTable[i]), nn))
    }

    if err != nil {
      logger.Debug("Error while reading segment data for packet in OggOpusReader", "error", err)
      return n, nil, err
    }
  }

  packet := &oggPage{
    headerType: oggHeaderType(headerTypeBuf[0]),
    granulePosition: binary.LittleEndian.Uint64(granulePositionBuf),
    bitstreamSerialNumber: binary.LittleEndian.Uint32(bitstreamSerialNumberBuf),
    pageSequenceNumber: binary.LittleEndian.Uint32(pageSequenceNumberBuf),
    checksum: binary.LittleEndian.Uint32(checksumBuf),
    segments: segmentData,
  }

  return n, packet, nil
  */
  panic("not implemented")
}
