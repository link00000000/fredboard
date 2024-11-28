package logging

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

var ErrNoCaller = errors.New("no caller")

func getCaller() (*runtime.Frame, error) {
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

type Level int

const (
	Debug = iota
	Info
	Warn
	Error
	Fatal
	Panic
)

type LoggerState int

const (
	LoggerState_Open = iota
	LoggerState_Closed
)

type OnLoggerCreatedEvent struct {
	Time   time.Time
	Caller *runtime.Frame
}

type OnLoggerClosedEvent struct {
	Time   time.Time
	Caller *runtime.Frame
}

type OnRecordEvent struct {
	Time    time.Time
	Level   Level
	Message string
	Error   error
	Caller  *runtime.Frame
}

type Handler interface {
	OnLoggerCreated(logger *Logger, event OnLoggerCreatedEvent)
	OnLoggerClosed(logger *Logger, event OnLoggerClosedEvent) error
	OnRecord(logger *Logger, event OnRecordEvent) error
}

type Logger struct {
	id       uuid.UUID
	parent   *Logger
	children []*Logger

	state LoggerState

	level        Level
	panicOnError bool
	handlers     []Handler
	data         map[string]any
}

func NewLogger() *Logger {
	return &Logger{
		id:       uuid.New(),
		children: make([]*Logger, 0),
		state:    LoggerState_Open,
		handlers: make([]Handler, 0),
		data:     make(map[string]any),
	}
}

func (logger *Logger) NewChildLogger() *Logger {
	childLogger := NewLogger()
	childLogger.parent = logger

	logger.children = append(logger.children, childLogger)

	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		panic(err)
	}

	event := OnLoggerCreatedEvent{
		Time:   time.Now().UTC(),
		Caller: caller,
	}

	for _, handler := range childLogger.Handlers() {
		handler.OnLoggerCreated(childLogger, event)
	}

	return childLogger
}

// Implements [io.Closer]
func (logger *Logger) Close() error {
	// Prevent closing a logger multiple times
	if logger.state == LoggerState_Closed {
		return nil
	}

	errs := make([]error, 0)

	for _, child := range logger.children {
		err := child.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		return err
	}

	event := OnLoggerClosedEvent{
		Time:   time.Now().UTC(),
		Caller: caller,
	}

	for _, handler := range logger.Handlers() {
		err := handler.OnLoggerClosed(logger, event)
		if err != nil {
			errs = append(errs, err)
		}
	}

	logger.state = LoggerState_Closed

	return errors.Join(errs...)
}

func (logger *Logger) RootLogger() *Logger {
	l := logger

	for l.parent != nil {
		l = l.parent
	}

	return l
}

func (logger *Logger) Handlers() []Handler {
	return logger.RootLogger().handlers
}

func (logger *Logger) AddHandler(handler Handler) {
	logger.RootLogger().handlers = append(logger.RootLogger().handlers, handler)
}

func (logger *Logger) Level() Level {
	return logger.RootLogger().level
}

func (logger *Logger) SetLevel(level Level) {
	logger.RootLogger().level = level
}

func (logger *Logger) PanicOnError() bool {
	return logger.RootLogger().panicOnError
}

func (logger *Logger) SetPanicOnError(value bool) {
	logger.RootLogger().panicOnError = value
}

func (logger *Logger) SetData(key string, value any) {
	logger.data[key] = value
}

var ErrInsignificantLevel = errors.New("insignificant log level")

func (logger *Logger) Log(message string, level Level, logError error) error {
	if level < logger.level {
		return ErrInsignificantLevel
	}

	caller, err := getCaller()

	// Ignore ErrNoCaller and continue to log without the caller
	if err != nil && err != ErrNoCaller {
		return err
	}

	event := OnRecordEvent{
		Time:    time.Now().UTC(),
		Level:   level,
		Message: message,
		Error:   logError,
		Caller:  caller,
	}

	errs := make([]error, 0)
	for _, handler := range logger.Handlers() {
		err := handler.OnRecord(logger, event)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (logger *Logger) Debug(message string) error {
	xerr := logger.Log(message, Debug, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) DebugWithErr(message string, err error) error {
	xerr := logger.Log(message, Debug, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) Info(message string) error {
	xerr := logger.Log(message, Info, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) InfoWithErr(message string, err error) error {
	xerr := logger.Log(message, Info, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) Warn(message string) error {
	xerr := logger.Log(message, Warn, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) WarnWithErr(message string, err error) error {
	xerr := logger.Log(message, Warn, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) Error(message string) error {
	xerr := logger.Log(message, Error, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) ErrorWithErr(message string, err error) error {
	xerr := logger.Log(message, Error, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	return xerr
}

func (logger *Logger) Fatal(message string) {
	xerr := logger.Log(message, Fatal, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	os.Exit(1)
}

func (logger *Logger) FatalWithErr(message string, err error) {
	xerr := logger.Log(message, Fatal, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	os.Exit(1)
}

func (logger *Logger) Panic(message string) {
	xerr := logger.Log(message, Panic, nil)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	panic("an unrecoverable error has occurred")
}

func (logger *Logger) PanicWithErr(message string, err error) {
	xerr := logger.Log(message, Panic, err)
	if logger.PanicOnError() && xerr != nil {
		panic(xerr)
	}

	panic("an unrecoverable error has occurred")
}
