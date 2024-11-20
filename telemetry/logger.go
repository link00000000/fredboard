package telemetry

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
)

var ErrInsignificant = errors.New("message was not processed because configured log level is too low")

var projectRoot string = "/"

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	projectRoot = filepath.Dir(filepath.Dir(thisFile))
}

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
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
	Message string     `json:"message"`
	Err     error      `json:"error"`
	Caller  *Caller    `json:"caller"`
	Context Context    `json:"context"`
}

func NewRecord(
	logger *Logger,
	time time.Time,
	level Level,
	message string,
	err error,
	caller *Caller,
	context Context,
) Record {
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
		Message: message,
		Err:     err,
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
  level    Level
}

func NewLogger(handlers []Handler) *Logger {
  logger := &Logger{id: uuid.New(), level: LevelInfo}
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

func (logger *Logger) SetLevel(level Level) {
  logger.level = level
}

func (logger *Logger) Log(level Level, message string, err error, context Context) error {
  if level > logger.level {
    return ErrInsignificant
  }

	var caller *Caller
	if pc, file, line, ok := runtime.Caller(1); ok {
		frames := runtime.CallersFrames([]uintptr{pc})
		frame, _ := frames.Next()

		file, _ := filepath.Rel(projectRoot, file)
		caller = &Caller{file, line, frame.Function}
	}

	errs := []error{}
	for _, handler := range logger.handlers {
		r := NewRecord(logger, time.Now(), level, message, err, caller, context)
		if err := handler.Handle(r); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (logger *Logger) FatalWithContext(message string, err error, context Context) {
	logger.Log(LevelFatal, message, err, context)
  os.Exit(1)
}

func (logger *Logger) Fatal(message string, err error) {
	logger.FatalWithContext(message, err, nil)
}

func (logger *Logger) ErrorWithContext(message string, err error, context Context) error {
	return logger.Log(LevelError, message, err, context)
}

func (logger *Logger) Error(message string, err error) error {
	return logger.ErrorWithContext(message, err, nil)
}

func (logger *Logger) WarnWithContext(message string, context Context) error {
	return logger.Log(LevelWarn, message, nil, context)
}

func (logger *Logger) Warn(message string) error {
	return logger.WarnWithContext(message, nil)
}

func (logger *Logger) InfoWithContext(message string, context Context) error {
	return logger.Log(LevelInfo, message, nil, context)
}

func (logger *Logger) Info(message string) error {
	return logger.InfoWithContext(message, nil)
}

func (logger *Logger) DebugWithContext(message string, context Context) error {
	return logger.Log(LevelDebug, message, nil, context)
}

func (logger *Logger) Debug(message string) error {
	return logger.DebugWithContext(message, nil)
}

func (logger *Logger) Err(err error) error {
	return logger.Error("an error has occurred", err)
}
