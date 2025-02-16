package logging

import (
	"fmt"
	"io"
)

type ReaderLogger struct {
	reader io.Reader
	logger *Logger
	level  Level
}

// Implements [io.Reader]
func (rl *ReaderLogger) Read(p []byte) (n int, err error) {
	n, err = rl.reader.Read(p)
	rl.logger.Log(rl.level, fmt.Sprintf("%s\n", p))

	return
}

func LogReader(reader io.Reader, logger *Logger, level Level) io.Reader {
	return &ReaderLogger{reader, logger, level}
}

type WriterLogger struct {
	writer io.Writer
	logger *Logger
	level  Level
}

// Implements [io.Writer]
func (wl *WriterLogger) Write(p []byte) (n int, err error) {
	n, err = wl.writer.Write(p)
	wl.logger.Log(wl.level, fmt.Sprintf("%s\n", p))

	return
}

func LogWriter(writer io.Writer, logger *Logger, level Level) io.Writer {
	return &WriterLogger{writer, logger, level}
}
