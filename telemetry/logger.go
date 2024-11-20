package telemetry

import (
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
)

var projectRoot string = "/"

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	projectRoot = filepath.Dir(filepath.Dir(thisFile))
}

const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

type Level uint8
type Context map[string]any

type Caller struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

type Record struct {
	Logger  uuid.UUID  `json:"logger"`
	Parent  *uuid.UUID `json:"parentLogger"`
	Root    uuid.UUID  `json:"rootLogger"`
	Time    time.Time  `json:"time"`
	Level   Level      `json:"level"`
	Caller  *Caller    `json:"caller"`
	Context Context    `json:"context"`
}

func NewRecord(logger *Logger, time time.Time, level Level, caller *Caller, context Context) Record {
	var parent *uuid.UUID
	if logger.parent != nil {
		parent = &logger.parent.id
	}

	return Record{
		Logger:  logger.id,
		Parent:  parent,
		Root:    logger.root.id,
		Time:    time,
		Level:   level,
		Caller:  caller,
		Context: context,
	}
}

type Handler interface {
	Handle(record Record) error
}

type JsonHandler struct {
	writer io.Writer
}

func NewJsonHandler(writer io.Writer) *JsonHandler {
	return &JsonHandler{writer}
}

func (handler *JsonHandler) Handle(record Record) error {
	data, err := json.Marshal(record)

	if err != nil {
		return err
	}

	_, err = handler.writer.Write(append(data, byte('\n')))
	return err
}

type Logger struct {
	id       uuid.UUID
	parent   *Logger
	root     *Logger
	handlers []Handler
}

func NewLogger(handlers []Handler) *Logger {
	logger := &Logger{id: uuid.New()}
	logger.root = logger
	logger.handlers = handlers[:]

	return logger
}

func NewLoggerWithParent(parent *Logger) *Logger {
	logger := NewLogger(parent.handlers)
	logger.parent = parent
	logger.root = parent.root

	return logger
}

// Implements [io.Closer]
func (logger *Logger) Close() error {
	errs := []error{}

	for _, handler := range logger.handlers {
		if handler, ok := handler.(io.Closer); ok {
			if err := handler.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (logger *Logger) Log(level Level, context Context) error {
	errs := []error{}

	var caller *Caller
	if pc, file, line, ok := runtime.Caller(1); ok {
		frames := runtime.CallersFrames([]uintptr{pc})
		frame, _ := frames.Next()

		file, _ := filepath.Rel(projectRoot, file)
		caller = &Caller{file, line, frame.Function}
	}

	for _, handler := range logger.handlers {
		r := NewRecord(logger, time.Now(), level, caller, context)
		handler.Handle(r)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
