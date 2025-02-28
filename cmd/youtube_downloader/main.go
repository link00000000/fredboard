package main

import (
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

	videoReader, err := ytdlp.NewVideoReader(
		logger,
		ytdlpConfig,
		"https://www.youtube.com/watch?v=F1oKhsy8wGw",
		ytdlp.YtdlpAudioQuality_BestAudio,
	)

	if err != nil {
		panic(err)
	}

	outputFile, err := os.Create("output.wav")
	if err != nil {
		panic(err)
	}

	defer outputFile.Close()

	readerNode := graph.NewReaderNode(logger, videoReader, 0x8000)
	writerNode := graph.NewWriterNode(logger, outputFile)

	audioGraph.AddNode(readerNode)
	audioGraph.AddNode(writerNode)
	audioGraph.CreateConnection(readerNode, writerNode)

	logger.Info("starting audio graph")

	for {
		audioGraph.Tick()

		if readerNode.Err() == io.EOF {
			break
		}

		if err := audioGraph.Err(); err != nil {
			panic(err)
		}
	}

	logger.Info("finished audio graph")
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
