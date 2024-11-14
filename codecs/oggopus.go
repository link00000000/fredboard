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

type OggHeaderType uint8

const (
	OggHeaderType_Continuation = 1 << iota
	OggHeaderType_BeginningOfStream
	OggHeaderType_EndOfStream
)

type OggPacket struct {
  HeaderType OggHeaderType
  GranulePosition uint64
  BitstreamSerialNumber uint32
  PageSequenceNumber uint32
  Checksum uint32
  Segments [][]byte
}

type OggOpusReader struct {
	internalReader io.Reader
}

func NewOggOpusReader(reader io.Reader) *OggOpusReader {
	return &OggOpusReader{ internalReader: reader }
}

func (oor *OggOpusReader) ReadNextPacket() (int, *OggPacket, error) {
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

    // TODO: Remove
    // Segment table can contain segments of length 0, skip it
    /*
    if segmentTable[i] == 0x00 {
      segmentData = append(segmentData, []byte{})
      continue
    }
    */

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

  packet := &OggPacket{
    HeaderType: OggHeaderType(headerTypeBuf[0]),
    GranulePosition: binary.LittleEndian.Uint64(granulePositionBuf),
    BitstreamSerialNumber: binary.LittleEndian.Uint32(bitstreamSerialNumberBuf),
    PageSequenceNumber: binary.LittleEndian.Uint32(pageSequenceNumberBuf),
    Checksum: binary.LittleEndian.Uint32(checksumBuf),
    Segments: segmentData,
  }

  return n, packet, nil
}
