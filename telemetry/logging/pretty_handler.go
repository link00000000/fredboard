package logging

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"accidentallycoded.com/fredboard/v3/ansi"
)

var projectRoot string = "/"

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	projectRoot = filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
}

type PrettyHandler struct {
	writer io.Writer
}

func NewPrettyHandler(writer io.Writer) PrettyHandler {
	return PrettyHandler{writer: writer}
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerCreated(logger *Logger, event OnLoggerCreatedEvent) error {
	return nil
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerClosed(logger *Logger, event OnLoggerClosedEvent) error {
	return nil
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnRecord(logger *Logger, event OnRecordEvent) error {
	var str strings.Builder

	str.WriteString(event.Time.Format("2006/01/02 15:04:05"))
	str.WriteString(" ")

	switch event.Level {
	case Debug:
		str.WriteString(ansi.FgMagenta + "DBG" + ansi.Reset)
	case Info:
		str.WriteString(ansi.FgBlue + "INF" + ansi.Reset)
	case Warn:
		str.WriteString(ansi.FgYellow + "WRN" + ansi.Reset)
	case Error:
		str.WriteString(ansi.FgRed + "ERR" + ansi.Reset)
	case Fatal:
		str.WriteString(ansi.FgBlack + ansi.BgRed + "FTL" + ansi.Reset)
	case Panic:
		str.WriteString(ansi.FgBlack + ansi.BgRed + "!!!" + ansi.Reset)
	}
	str.WriteString(" ")

	var callerRelativePath *string
	if event.Caller != nil {
		if relativePath, err := filepath.Rel(projectRoot, event.Caller.File); err == nil {
			callerRelativePath = &relativePath
		}
	}

	if callerRelativePath != nil {
		str.WriteString(ansi.FgBrightBlack + fmt.Sprintf("<%s:%d>", *callerRelativePath, event.Caller.Line) + ansi.Reset)
	} else {
		str.WriteString(ansi.FgBrightBlack + "<UNKNOWN CALLER>" + ansi.Reset)
	}
	str.WriteString(" ")

	str.WriteString(event.Message)
	str.WriteString("\n")

	if event.Error != nil {
		if len(logger.data) != 0 {
			str.WriteString("                     ├─ ")
		} else {
			str.WriteString("                     └─ ")
		}

		str.WriteString(ansi.FgRed + event.Error.Error() + ansi.Reset + "\n")
	}

	i := 0
	for k, v := range logger.data {
		if i < len(logger.data)-1 {
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
