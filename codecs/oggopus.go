package codecs

import (
	"encoding/binary"
	"errors"
	"fmt"
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

type oggPageHeader struct {
  headerType oggHeaderType
  granulePosition uint64
  bitstreamSerialNumber uint32
  pageSequenceNumber uint32
  checksum uint32
  segmentTable []uint8
}

type oggPageDataSegment []byte

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
  state oggReaderState
  currentPageHeader *oggPageHeader
  currentDataSegmentIdx int
}

func NewOggOpusReader(reader io.Reader) *OggOpusReader {
	return &OggOpusReader{
    internalReader: reader,
    state: oggReaderState_ready,
    currentPageHeader: nil,
    currentDataSegmentIdx: 0,
  }
}

func (r *OggOpusReader) ReadNextOpusPacket() (int, OpusPacket, error) {
  n := 0

  for {
    switch r.state {
    case oggReaderState_ready:
      logger.Debug("ReadNextOggPacket: beginning of file", "reader", r)
      r.state = oggReaderState_readingHeader

    case oggReaderState_readingHeader:
      nn, pageHeader, err := readOggPageHeader(r.internalReader)
      n += nn

      logger.Debug("ReadNextOggPacket: read page header", "reader", r, "numBytesRead", nn, "pageHeader", pageHeader)

      if err == io.EOF {
        logger.Debug("ReadNextOggPacket: end of file", "reader", r)
        r.state = oggReaderState_done
        break
      }

      if err != nil {
        logger.Error("ReadNextOggPacket: error while reading most recent page header", "reader", r, "error", err)
        r.state = oggReaderState_failed
        return n, nil, err
      }

      r.currentPageHeader = pageHeader
      r.currentDataSegmentIdx = 0

      logger.Debug("ReadNextOggPacket: set current page header", "reader", r, "pageHeader", pageHeader)

      r.state = oggReaderState_readingData

    case oggReaderState_readingData:
      if (r.currentPageHeader == nil) {
        panic("Tried to read page data before reading page header")
      }

      if (r.currentDataSegmentIdx >= len(r.currentPageHeader.segmentTable)) {
        logger.Debug("ReadNextOggPacket: end of page")
        r.state = oggReaderState_readingHeader
        break
      }

      nn, pageDataSegment, err := readOggPageDataSegment(r.internalReader, r.currentPageHeader.segmentTable[r.currentDataSegmentIdx])
      n += nn

      logger.Debug("ReadNextOggPacket: read page data segment", "reader", r, "numBytesRead", nn, "pageDataSegment", pageDataSegment)

      if err != nil {
        logger.Error("ReadNextOggPacket: error while reading most recent page data segment", "reader", r, "error", err)
        r.state = oggReaderState_failed
        return n, nil, err
      }

      // If the segment indicates the start of a header packet, skip the entire page
      x := r.currentDataSegmentIdx == 0
      y := isStartOfAHeaderPacket(pageDataSegment)

      if x && y {
        for i := 1; i < len(r.currentPageHeader.segmentTable); i++ {
          nn, _, err := readOggPageDataSegment(r.internalReader, r.currentPageHeader.segmentTable[i])
          n += nn

          logger.Debug("ReadNextOggPacket: skipped data segment", "numBytes", nn)

          if err != nil {
            logger.Error("ReadNextOggPacket: error while skipping most recent data segment", "error", err)
            return n, nil, err
          }

        }

        r.state = oggReaderState_readingHeader
        break
      }

      r.currentDataSegmentIdx++
      logger.Debug("ReadNextOggPacket: advancing current data segment index", "reader", r)

      return n, OpusPacket(*pageDataSegment), nil
    
    case oggReaderState_done:
      return 0, nil, io.EOF

    case oggReaderState_failed:
      return 0, nil, errors.New("Attempted to call ReadNextOggPacket() on failed reader")
    }
  }
}

func isStartOfAHeaderPacket(pageDataSegment *oggPageDataSegment) bool {
  seg := []byte(*pageDataSegment)
  return len(seg) >= 2 && string(seg[:2]) == "Op"
}

func readOggPageHeader(r io.Reader) (int, *oggPageHeader, error) {
	n := 0

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

  numPageSegmentsBuf := make([]byte, 1)
  nn, err = r.Read(numPageSegmentsBuf)
  n += n

  if err != nil && nn != 1 {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for page segments, got %d bytes", 1, nn))
  }

	if err != nil {
		logger.Debug("Error while reading page segments for packet in OggOpusReader", "error", err)
		return n, nil, err
	}

	logger.Debug("Read page segments for packet in OggOpusReader", "pageSegments", numPageSegmentsBuf)

  segmentTable := make([]byte, int(numPageSegmentsBuf[0]))
  nn, err = r.Read(segmentTable)
  n += nn

  if err != nil && nn != int(numPageSegmentsBuf[0]) {
		err = errors.New(fmt.Sprintf("Expected to read %d bytes for segment table, got %d bytes", int(numPageSegmentsBuf[0]), nn))
  }

  if err != nil {
		logger.Debug("Error while reading segment table for packet in OggOpusReader", "error", err)
		return n, nil, err
  }

  header := &oggPageHeader{
    headerType: oggHeaderType(headerTypeBuf[0]),
    granulePosition: binary.LittleEndian.Uint64(granulePositionBuf),
    bitstreamSerialNumber: binary.LittleEndian.Uint32(bitstreamSerialNumberBuf),
    pageSequenceNumber: binary.LittleEndian.Uint32(pageSequenceNumberBuf),
    checksum: binary.LittleEndian.Uint32(checksumBuf),
    segmentTable: segmentTable,
  };

  return n, header, nil
}

func readOggPageDataSegment(r io.Reader, segmentLength uint8) (int, *oggPageDataSegment, error) {
  segmentData := make([]byte, segmentLength)
  n, err := r.Read(segmentData)

  if err != nil {
    logger.Error("readOggDataSegment: Failed to next segment", "reader", r, "segmentLength", segmentLength, "error", err)
    return n, nil, err
  }

  logger.Debug("readOggDataSegment: Read next segment", "reader", r, "segmentLength", segmentLength, "numBytesRead", n, "segmentData", segmentData)

  return n, (*oggPageDataSegment)(&segmentData), nil
}
