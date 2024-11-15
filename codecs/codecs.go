package codecs

type OpusPacket []byte

type Reader interface {
  ReadNextOpusPacket() (int, OpusPacket, error)
}
