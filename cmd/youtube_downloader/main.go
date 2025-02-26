package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	_ "accidentallycoded.com/fredboard/v3/internal/telemetry/pprof"
)

const url = "https://www.youtube.com/watch?v=OkktfeAR-Rk"

var (
	logger *logging.Logger

	ytdlpConfig  *ytdlp.Config
	ffmpegConfig *ffmpeg.Config
)

func init() {
	initializeConfig()

	logger = initializeLogger(config.Get())

	ytdlpConfig = &ytdlp.Config{
		ExePath:     optional.MakePtr(config.Get().Ytdlp.ExePath),
		CookiesPath: optional.MakePtr(config.Get().Ytdlp.CookiesFile),
	}

	ffmpegConfig = &ffmpeg.Config{
		ExePath: optional.MakePtr(config.Get().Ffmpeg.ExePath),
	}
}

func main() {
	audioGraph := graph.NewGraph(logger)

	input := bytes.NewBufferString("1-2-3-4-5-6-7-8-9-10-11-12-13-14-15-16-1-2-3-4-5-6-7-8-9-10-11-12-13-14-15-16")
	var output bytes.Buffer

	readerNode := graph.NewReaderNode(logger, input, 3)
	writerNode := graph.NewWriterNode(logger, &output)

	audioGraph.AddNode(readerNode)
	audioGraph.AddNode(writerNode)
	audioGraph.CreateConnection(readerNode, writerNode)

	logger.Info("starting audio graph", "data", input.String())

	for {
		audioGraph.Tick()

		if readerNode.Err() == io.EOF {
			break
		}

		if err := audioGraph.Err(); err != nil {
			panic(err)
		}
	}

	logger.Info("finished audio graph", "data", output.String())
}

func initializeConfig() {
	if err := config.Init(); err != nil {
		fmt.Printf("failed to initialize config: %s", err.Error())
		os.Exit(1)
	}

	if ok, errs := config.Validate(); !ok {
		fmt.Printf("invalid config:\n%s", errors.Join(errs...))
		os.Exit(1)
	}
}

func initializeLogger(settings config.Settings) *logging.Logger {
	logger := logging.NewLogger()
	logger.SetPanicOnError(true)

	for _, handlerConfig := range settings.Loggers.Handlers {
		var w io.Writer
		if *handlerConfig.Output == "stdout" {
			w = os.Stdout
		} else if *handlerConfig.Output == "stderr" {
			w = os.Stderr
		} else {
			f, err := os.Open(*handlerConfig.Output)

			if err != nil {
				fmt.Printf("failed to create logger: %s", err.Error())
				os.Exit(1)
			}

			defer f.Close()
			w = f
		}

		var handler logging.Handler
		switch *handlerConfig.Type {
		case config.LoggingHandlerType_Pretty:
			handler = logging.NewPrettyHandler(w, *handlerConfig.Level)
		case config.LoggingHandlerType_JSON:
			handler = logging.NewJsonHandler(w, *handlerConfig.Level)
		}

		logger.AddHandler(handler)
	}

	return logger
}
