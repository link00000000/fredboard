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
type Data map[string]any

type Context struct {
	id       uuid.UUID
	logger   *Logger
	children []*Context
	parent   *Context
	data     Data
}

func (logger *Logger) NewContext(parent *Context) *Context {
	return &Context{
		id:     uuid.New(),
		logger: logger,
		parent: parent,
		data:   make(Data),
	}
}

func (ctx *Context) SetValue(name string, value any) {
	ctx.data[name] = value
}

// Implements [io.Closer]
func (ctx *Context) Close() error {
	return ctx.logger.onContextClosed(*ctx)
}

type Record struct {
	Time    time.Time
	Level   Level
	Message string
	Err     error
	Caller  *runtime.Frame
	Context Context
}

func NewRecord(
	time time.Time,
	level Level,
	message string,
	err error,
	caller *runtime.Frame,
	ctx Context,
) Record {
	return Record{
		Time:    time,
		Level:   level,
		Message: message,
		Err:     err,
		Caller:  caller,
		Context: ctx,
	}
}

type Handler interface {
	OnRecord(record Record) error
	OnContextClosed(ctx Context) error
}

type JsonHandlerRecord struct {
	Time    time.Time          `json:"time"`
	Level   Level              `json:"level"`
	Message string             `json:"message"`
	Err     error              `json:"error"`
	Caller  *runtime.Frame     `json:"caller"`
	Context JsonHandlerContext `json:"context"`
}

func NewJsonHandlerRecord(record Record) JsonHandlerRecord {
	return JsonHandlerRecord{
		Time:    record.Time,
		Level:   record.Level,
		Message: record.Message,
		Err:     record.Err,
		Caller:  record.Caller,
		Context: NewJsonHandlerContext(record.Context),
	}
}

type JsonHandlerContext struct {
	Id       uuid.UUID   `json:"id"`
	Children []uuid.UUID `json:"children"`
	Parent   *uuid.UUID  `json:"parent"`
	Data     Data        `json:"data"`
}

func NewJsonHandlerContext(ctx Context) JsonHandlerContext {
	children := make([]uuid.UUID, len(ctx.children))
	for i, c := range ctx.children {
		children[i] = c.id
	}

	newCtx := JsonHandlerContext{
		Id:       ctx.id,
		Children: children,
		Data:     ctx.data,
	}

	if ctx.parent != nil {
		newCtx.Parent = &ctx.parent.id
	}

	return newCtx
}

type JsonHandler struct {
	writer io.Writer
}

func NewJsonHandler(writer io.Writer) *JsonHandler {
	return &JsonHandler{writer}
}

func (handler *JsonHandler) OnRecord(record Record) error {
	data, err := json.Marshal(NewJsonHandlerRecord(record))

	if err != nil {
		return err
	}

	_, err = handler.writer.Write(append(data, byte('\n')))
	return err
}

func (handler *JsonHandler) OnContextClosed(ctx Context) error {
	return nil
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

func (handler *PrettyHandler) OnRecord(record Record) error {
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
		if len(record.Context.data) != 0 {
			str.WriteString("                     ├─ ")
		} else {
			str.WriteString("                     └─ ")
		}

		str.WriteString(ansi.FgRed + record.Err.Error() + ansi.Reset + "\n")
	}

	i := 0
	for k, v := range record.Context.data {
		if i < len(record.Context.data)-1 {
			str.WriteString("                     ├─ ")
		} else {
			str.WriteString("                     └─ ")
		}

		i++

		str.WriteString(fmt.Sprintf("%s: %#v", ansi.FgBrightBlack+k+ansi.Reset, v))
		str.WriteString("\n")
	}

	_, err := fmt.Fprintf(handler.writer, str.String())
	return err
}

func (*PrettyHandler) OnContextClosed(ctx Context) error {
	return nil
}

type Logger struct {
	handlers []Handler
	level    Level
	RootCtx  *Context
}

func NewLogger(handlers []Handler) *Logger {
	logger := &Logger{level: LevelInfo}
	logger.RootCtx = logger.NewContext(nil)

	if handlers != nil {
		logger.handlers = handlers[:]
	}

	return logger
}

func (logger *Logger) AddHandler(handler Handler) {
	logger.handlers = append(logger.handlers, handler)
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
		r := NewRecord(time.Now(), level, message, loggedErr, caller, context)
		if err := handler.OnRecord(r); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (logger *Logger) Fatal(message string, err error, ctx *Context) {
	logger.Log(LevelFatal, message, err, *ctx)
	os.Exit(1)
}

func (logger *Logger) Error(message string, err error, ctx *Context) error {
	return logger.Log(LevelError, message, err, *ctx)
}

func (logger *Logger) Warn(message string, ctx *Context) error {
	return logger.Log(LevelWarn, message, nil, *ctx)
}

func (logger *Logger) Info(message string, ctx *Context) error {
	return logger.Log(LevelInfo, message, nil, *ctx)
}

func (logger *Logger) Debug(message string, ctx *Context) error {
	return logger.Log(LevelDebug, message, nil, *ctx)
}

func (logger *Logger) Err(err error, ctx *Context) error {
	return logger.Error("an error has occurred", err, ctx)
}

func (logger *Logger) onContextClosed(ctx Context) error {
	errs := []error{}
	for _, handler := range logger.handlers {
		handler.OnContextClosed(ctx)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
