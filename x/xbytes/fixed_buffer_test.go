package xbytes

import (
	"io"
	"testing"
)

func eqSlices[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func TestSimpleWriteRead(t *testing.T) {
	buf := NewFixedBuffer(8)
	n, err := buf.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 8 {
		t.Fatalf("expected to write 8 bytes, wrote %d", n)
	}

	result := make([]byte, 8)
	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 8 {
		t.Fatalf("expected to read 8 bytes, read %d", n)
	}

	if !eqSlices(result, []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("incorrect result vaule: %v", result)
	}
}

func TestMultipleWrites(t *testing.T) {
	buf := NewFixedBuffer(8)
	n, err := buf.Write([]byte{1, 2, 3, 4})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to write 4 bytes, wrote %d", n)
	}

	n, err = buf.Write([]byte{5, 6, 7, 8})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to write 4 bytes, wrote %d", n)
	}

	result := make([]byte, 8)
	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 8 {
		t.Fatalf("expected to read 8 bytes, read %d", n)
	}

	if !eqSlices(result, []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("incorrect result vaule: %v", result)
	}
}

func TestMultipleReads(t *testing.T) {
	buf := NewFixedBuffer(8)
	n, err := buf.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 8 {
		t.Fatalf("expected to write 8 bytes, wrote %d", n)
	}

	result := make([]byte, 4)
	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to read 4 bytes, read %d", n)
	}

	if !eqSlices(result, []byte{1, 2, 3, 4}) {
		t.Fatalf("incorrect result vaule: %v", result)
	}

	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to read 4 bytes, read %d", n)
	}

	if !eqSlices(result, []byte{5, 6, 7, 8}) {
		t.Fatalf("incorrect result vaule: %v", result)
	}
}

func TestWriteExceedsBufferCapacity(t *testing.T) {
	buf := NewFixedBuffer(4)
	n, err := buf.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to write 4 byte, wrote %d", n)
	}

	n, err = buf.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	if err != ErrBufferFull {
		t.Fatalf("expected error %v, but got %v", ErrBufferFull, err)
	}

	if n != 0 {
		t.Fatalf("expected to write 0 byte, wrote %d", n)
	}

	result := make([]byte, 4)
	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to read 4 bytes, read %d", n)
	}

	if !eqSlices(result, []byte{1, 2, 3, 4}) {
		t.Fatalf("incorrect result vaule: %v", result)
	}
}

func TestReadEmptyBuffer(t *testing.T) {
	buf := NewFixedBuffer(4)

	n, err := buf.Write([]byte{1, 2, 3, 4})

	if err != nil {
		t.Fatalf("error while calling buf.Write: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to write 4 byte, wrote %d", n)
	}

	result := make([]byte, 4)
	n, err = buf.Read(result)

	if err != nil {
		t.Fatalf("error while calling buf.Read: %v", err)
	}

	if n != 4 {
		t.Fatalf("expected to read 4 byte, read %d", n)
	}

	n, err = buf.Read(result)

	if err != io.EOF {
		t.Fatalf("expected error %v while calling buf.Read, got %v", io.EOF, err)
	}

	if n != 0 {
		t.Fatalf("expected to read 0 byte, read %d", n)
	}
}
