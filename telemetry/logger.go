package telemetry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"accidentallycoded.com/fredboard/v3/ansi"
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

type Record struct {
	Logger  uuid.UUID      `json:"logger"`
	Parent  *uuid.UUID     `json:"parentLogger"`
	Root    uuid.UUID      `json:"rootLogger"`
	Time    time.Time      `json:"time"`
	Level   Level          `json:"level"`
	Message string         `json:"message"`
	Err     error          `json:"error"`
	Caller  *runtime.Frame `json:"caller"`
	Context Context        `json:"context"`
}

func NewRecord(
	logger *Logger,
	time time.Time,
	level Level,
	message string,
	err error,
	caller *runtime.Frame,
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

var ErrNoCaller = errors.New("no caller")

func getModulePath(functionPath string) string {
	// Module paths contain parens for the struct and have an additional dot in the path
	// ex.
	//  Module: accidentallycoded.com/fredboard/v3/telemetry.(*Logger).Log
	//  Function: ccidentallycoded.com/fredboard/v3/telemetry.NewLogger
	isMethod := strings.Contains(functionPath, "(")

	var endOfModuleName int
	if isMethod {
		endOfModuleName = strings.LastIndex(functionPath, "(") - 1
	} else {
		endOfModuleName = strings.LastIndex(functionPath, ".")
	}

	if endOfModuleName == -1 {
		panic("malformed module name does not contain a . delimiter")
	}

	return functionPath[:endOfModuleName]
}

func getModuleCallerFrame() (*runtime.Frame, error) {
	pcs := make([]uintptr, 8)
	n := runtime.Callers(1, pcs)
	pcs = pcs[:n]

	if len(pcs) == 0 {
		return nil, ErrNoCaller
	}

	frames := runtime.CallersFrames(pcs)

	firstFrame, more := frames.Next()
	if !more {
		return nil, ErrNoCaller
	}

	thisModule := getModulePath(firstFrame.Function)

	for {
		frame, more := frames.Next()
		module := getModulePath(frame.Function)

		if module != thisModule {
			return &frame, nil
		}

		if !more {
			break
		}
	}

	return nil, ErrNoCaller
}

type PrettyHandler struct {
	writer io.Writer
}

func NewPrettyHandler(writer io.Writer) *PrettyHandler {
	return &PrettyHandler{writer}
}

func (handler *PrettyHandler) Handle(record Record) error {
	var str strings.Builder

	str.WriteString(record.Time.Format("2006/01/02 15:04:05"))
	str.WriteString(" ")

	switch record.Level {
	case LevelDebug:
		str.WriteString(ansi.FgMagenta + "DBG" + ansi.Reset)
	case LevelInfo:
		str.WriteString(ansi.FgBlue + "INF" + ansi.Reset)
	case LevelWarn:
		str.WriteString(ansi.FgYellow + "WRN" + ansi.Reset)
	case LevelError:
		str.WriteString(ansi.FgRed + "ERR" + ansi.Reset)
	case LevelFatal:
		str.WriteString(ansi.FgBlack + ansi.BgRed + "FTL" + ansi.Reset)
	}
	str.WriteString(" ")

  var callerRelativePath *string
	if record.Caller != nil {
    if relativePath, err := filepath.Rel(projectRoot, record.Caller.File); err == nil {
      callerRelativePath = &relativePath
    }
  }
    
  if callerRelativePath != nil {
		str.WriteString(ansi.FgBrightBlack + fmt.Sprintf("<%s:%d>", *callerRelativePath, record.Caller.Line) + ansi.Reset)
	} else {
		str.WriteString(ansi.FgBrightBlack + "<UNKNOWN CALLER>" + ansi.Reset)
	}
	str.WriteString(" ")

	str.WriteString(record.Message)
	str.WriteString("\n")

  if record.Err != nil {
    if record.Context != nil && len(record.Context) != 0 {
      str.WriteString("                     ├─ ")
    } else {
      str.WriteString("                     └─ ")
    }

    str.WriteString(ansi.FgRed + record.Err.Error() + ansi.Reset + "\n")
  }

  i := 0
  for k, v := range record.Context {
    if i < len(record.Context) - 1 {
      str.WriteString("                     ├─ ")
    } else {
      str.WriteString("                     └─ ")
    }

    i++

    str.WriteString(fmt.Sprintf("%s: %#v", ansi.FgBrightBlack + k + ansi.Reset, v))
    str.WriteString("\n")
  }

  _, err := fmt.Fprintf(handler.writer, str.String())
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
  loggedErr := err

	if level > logger.level {
		return ErrInsignificant
	}

	caller, err := getModuleCallerFrame()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		return err
	}

	errs := []error{}
	for _, handler := range logger.handlers {
		r := NewRecord(logger, time.Now(), level, message, loggedErr, caller, context)
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
