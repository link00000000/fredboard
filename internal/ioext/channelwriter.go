package ioext

type ChannelWriter[T any] struct {
	c chan<- T
}

func (w *ChannelWriter[T]) Write(p []T) (n int, err error) {
	for _, q := range p {
		w.c <- q
	}

	return len(p), err
}

func (w *ChannelWriter[T]) Close() error {
	close(w.c)
	return nil
}

func NewChannelWriter[T any](c chan<- T) *ChannelWriter[T] {
	return &ChannelWriter[T]{c: c}
}
