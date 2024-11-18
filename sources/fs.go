package sources

import (
	"os"
)

type FS struct {
	f *os.File
}

func NewFSSource(filePath string) (*FS, error) {
	f, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	return &FS{f: f}, nil
}

// Implements [io.Closer]
func (fs *FS) Close() error {
	return fs.f.Close()
}

// Implements [io.Reader]
func (fs *FS) Read(p []byte) (int, error) {
	return fs.f.Read(p)
}
