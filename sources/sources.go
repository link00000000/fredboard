package sources

import "io"

type Source interface {
	io.Reader
	Start() error
	Wait() error
}
