package sources

import "io"

type Source interface {
	io.Reader
	Start() error
	Stop() error
	Wait() error
}
