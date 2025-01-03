package logging

import (
	"encoding/json"
	"io"
	"time"
)

type JsonHandlerMessageType int

const (
	JsonHandlerMessageType_LoggerCreated JsonHandlerMessageType = iota
	JsonHandlerMessageType_LoggerClosed
	JsonHandlerMessageType_Record
)

type JsonHandlerCaller struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type JsonHandlerLogger struct {
	Id       string   `json:"id"`
	Parent   *string  `json:"parent"`
	Children []string `json:"children"`
	Root     string   `json:"root"`
}

type JsonHandlerLoggerCreated struct {
	Time   time.Time         `json:"time"`
	Caller JsonHandlerCaller `json:"caller"`
	Logger JsonHandlerLogger `json:"logger"`
}

type JsonHandlerLoggerClosed struct {
	Time   time.Time         `json:"time"`
	Caller JsonHandlerCaller `json:"caller"`
	Logger JsonHandlerLogger `json:"logger"`
}

type JsonHandlerRecord struct {
	Time    time.Time         `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Error   *string           `json:"error"`
	Caller  JsonHandlerCaller `json:"caller"`
	Logger  JsonHandlerLogger `json:"logger"`
}

type JsonHandlerMessage[T any] struct {
	Type JsonHandlerMessageType `json:"type"`
	Data T                      `json:"data"`
}

func NewJsonLoggerCreatedMessage() JsonHandlerMessage[JsonHandlerLoggerCreated] {
	return JsonHandlerMessage[JsonHandlerLoggerCreated]{Type: JsonHandlerMessageType_LoggerCreated, Data: JsonHandlerLoggerCreated{}}
}

func NewJsonLoggerClosedMessage() JsonHandlerMessage[JsonHandlerLoggerClosed] {
	return JsonHandlerMessage[JsonHandlerLoggerClosed]{Type: JsonHandlerMessageType_LoggerClosed, Data: JsonHandlerLoggerClosed{}}
}

func NewJsonLoggerRecordMessage() JsonHandlerMessage[JsonHandlerRecord] {
	return JsonHandlerMessage[JsonHandlerRecord]{Type: JsonHandlerMessageType_Record, Data: JsonHandlerRecord{}}
}

type JsonHandler struct {
	writer io.Writer
}

func NewJsonHandler(writer io.Writer) JsonHandler {
	return JsonHandler{writer: writer}
}

// Implements [logging.Handler]
func (handler JsonHandler) OnLoggerCreated(logger *Logger, event OnLoggerCreatedEvent) {
	loggerCreated := NewJsonLoggerCreatedMessage()
	loggerCreated.Data.Time = event.Time

	loggerCreated.Data.Caller = JsonHandlerCaller{}
	loggerCreated.Data.Caller.File = event.Caller.File
	loggerCreated.Data.Caller.Line = event.Caller.Line

	loggerCreated.Data.Logger.Id = logger.id.String()
	loggerCreated.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		loggerCreated.Data.Logger.Parent = &str
	}

	loggerCreated.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		loggerCreated.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(loggerCreated)
	if err != nil {
		panic(err)
	}

	// Handle error?
	handler.writer.Write(append(data, byte('\n')))
}

// Implements [logging.Handler]
func (handler JsonHandler) OnLoggerClosed(logger *Logger, event OnLoggerClosedEvent) error {
	loggerClosed := NewJsonLoggerClosedMessage()
	loggerClosed.Data.Time = event.Time

	loggerClosed.Data.Caller = JsonHandlerCaller{}
	loggerClosed.Data.Caller.File = event.Caller.File
	loggerClosed.Data.Caller.Line = event.Caller.Line

	loggerClosed.Data.Logger.Id = logger.id.String()
	loggerClosed.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		loggerClosed.Data.Logger.Parent = &str
	}

	loggerClosed.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		loggerClosed.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(loggerClosed)
	if err != nil {
		return err
	}

	handler.writer.Write(append(data, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

// Implements [logging.Handler]
func (handler JsonHandler) OnRecord(logger *Logger, event OnRecordEvent) error {
	record := NewJsonLoggerRecordMessage()
	record.Data.Time = event.Time

	switch event.Level {
	case Debug:
		record.Data.Level = "debug"
	case Info:
		record.Data.Level = "info"
	case Warn:
		record.Data.Level = "warn"
	case Error:
		record.Data.Level = "error"
	case Fatal:
		record.Data.Level = "fatal"
	case Panic:
		record.Data.Level = "panic"
	}

	record.Data.Message = event.Message

	if event.Error != nil {
		msg := event.Error.Error()
		record.Data.Error = &msg
	}

	record.Data.Caller = JsonHandlerCaller{}
	record.Data.Caller.File = event.Caller.File
	record.Data.Caller.Line = event.Caller.Line

	record.Data.Logger.Id = logger.id.String()
	record.Data.Logger.Root = logger.RootLogger().id.String()

	if logger.parent != nil {
		str := logger.parent.id.String()
		record.Data.Logger.Parent = &str
	}

	record.Data.Logger.Children = make([]string, len(logger.children))
	for i, c := range logger.children {
		record.Data.Logger.Children[i] = c.id.String()
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	handler.writer.Write(append(data, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}
