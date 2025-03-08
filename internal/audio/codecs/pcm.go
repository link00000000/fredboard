package codecs

import "encoding/binary"

// convert little endian bytes to signed 16bit values
func BytesToS16LE(bytes []byte) (s16le []int16) {
	// TODO: handle len(bytes) not divisible by 2. if its not divisible by 2, the last byte will be dropped
	s16le = make([]int16, len(bytes)/2)

	for i := 0; i < len(bytes)/2; i = i + 1 {
		s16le[i] = int16(binary.LittleEndian.Uint16(bytes[i*2 : (i+1)*2]))
	}

	return s16le
}

// convert signed 16bit values to bytes represented in little endian
func S16LEToBytes(s16le []int16) (bytes []byte) {
	bytes = make([]byte, 0, len(s16le)*2)

	for _, v := range s16le {
		bytes = binary.LittleEndian.AppendUint16(bytes, uint16(v))
	}

	return bytes
}
