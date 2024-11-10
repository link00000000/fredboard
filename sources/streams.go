package sources

type OpusAudioStream interface {
	GetStreamChannel() <-chan []byte
	Start() error
	Close() error
	Wait()
}
