package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"accidentallycoded.com/fredboard/v3/ansi"
	"golang.org/x/term"
)

const globalPadding = "                     "

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

func (handler PrettyHandler) useColor() bool {
	file, ok := handler.writer.(*os.File)
	if !ok {
		return false
	}

	isTerm := term.IsTerminal(int(file.Fd()))
	return isTerm
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerCreated(logger *Logger, event OnLoggerCreatedEvent) {
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnLoggerClosed(logger *Logger, event OnLoggerClosedEvent) error {
	return nil
}

// Implements [logging.Handler]
func (handler PrettyHandler) OnRecord(logger *Logger, event OnRecordEvent) error {
	var str ansi.AnsiStringBuilder
	if handler.useColor() {
		str.SetEscapeMode(ansi.EscapeMode_Enable)
	} else {
		str.SetEscapeMode(ansi.EscapeMode_Disable)
	}

	str.Write(event.Time.Format("2006/01/02 15:04:05"), " ")

	switch event.Level {
	case Debug:
		str.Write(ansi.FgMagenta, "DBG", ansi.Reset)
	case Info:
		str.Write(ansi.FgBlue, "INF", ansi.Reset)
	case Warn:
		str.Write(ansi.FgYellow, "WRN", ansi.Reset)
	case Error:
		str.Write(ansi.FgRed, "ERR", ansi.Reset)
	case Fatal:
		str.Write(ansi.FgBlack, ansi.BgRed, "FTL", ansi.Reset)
	case Panic:
		str.Write(ansi.FgBlack, ansi.BgRed, "!!!", ansi.Reset)
	}

	str.WriteString(" ")

	var callerRelativePath *string
	if event.Caller != nil {
		if relativePath, err := filepath.Rel(projectRoot, event.Caller.File); err == nil {
			callerRelativePath = &relativePath
		}
	}

	if callerRelativePath != nil {
		str.Write(ansi.FgBrightBlack, fmt.Sprintf("<%s:%d> ", *callerRelativePath, event.Caller.Line), ansi.Reset)
	} else {
		str.Write(ansi.FgBrightBlack, "<UNKNOWN CALLER> ", ansi.Reset)
	}

	str.WriteString(event.Message)

	str.WriteString("\n")

	if event.Error != nil {
		if len(logger.data) != 0 {
			str.WriteString(globalPadding)
			str.WriteString("├─ ")
		} else {
			str.WriteString(globalPadding)
			str.WriteString("└─ ")
		}

		str.Write(ansi.FgRed, event.Error.Error(), ansi.Reset, "\n")
	}

	dataJson, err := json.Marshal(logger.data)
	if err != nil {
		return err
	}

	var dataMap map[string]any
	err = json.Unmarshal([]byte(dataJson), &dataMap)
	if err != nil {
		return err
	}

	printData(&str, dataMap, globalPadding)

	_, err = fmt.Fprintf(handler.writer, str.String())
	return err
}

func printData(str *ansi.AnsiStringBuilder, data map[string]any, padding string) {
	i := 0
	for k, v := range data {
		str.WriteString(padding)

		isLast := i == len(data)-1
		if !isLast {
			str.WriteString("├─ ")
		} else {
			str.WriteString("└─ ")
		}

		switch v := v.(type) {
		case map[string]any:
			str.Write(ansi.FgBrightBlack, k, ansi.Reset, "\n")

			if !isLast {
				printData(str, v, padding+"│   ")
			} else {
				printData(str, v, padding+"    ")
			}
		default:
			str.Write(ansi.FgBrightBlack, k, ansi.Reset, ": ", fmt.Sprintf("%#v", v), "\n")
		}

		i++
	}
}
