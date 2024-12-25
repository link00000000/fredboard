package sources

import (
	"os"
)

type FS struct {
	filePath string
	f        *os.File

	stopped bool

	done chan struct{}
}

func NewFSSource(filePath string) *FS {
	return &FS{filePath: filePath, stopped: false, done: make(chan struct{})}
}

// Implements [Source]
func (fs *FS) Read(p []byte) (int, error) {
	n, err := fs.f.Read(p)

	if err != nil {
		close(fs.done)
		return n, err
	}

	return n, err
}

// Implements [Source]
func (fs *FS) Start() error {
	f, err := os.Open(fs.filePath)

	if err != nil {
		return err
	}

	fs.f = f

	return nil
}

// Implements [Source]
func (fs *FS) Stop() error {
	if !fs.stopped {
		err := fs.f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Implements [Source]
func (fs *FS) Wait() error {
	<-fs.done

	err := fs.Stop()
	if err != nil {
		return err
	}

	return nil
}
