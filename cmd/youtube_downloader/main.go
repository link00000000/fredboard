package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	_ "accidentallycoded.com/fredboard/v3/internal/telemetry/pprof"
)

const url = "https://www.youtube.com/watch?v=OkktfeAR-Rk"

var logger *logging.Logger

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get current working directory: %s", err.Error()))
	}

	configFile := path.Join(cwd, "config.json")

	if envFredboardConfig, ok := os.LookupEnv("FREDBOARD_CONFIG"); ok {
		configFile = envFredboardConfig
	}

	verrs, err := config.Init(config.ConfigInitOptions{Files: []string{configFile}})
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config: %s", err.Error()))
	}

	if len(verrs) > 0 {
		for _, verr := range verrs {
			fmt.Printf("configuration validation failed: %s", verr.Error())
		}
	}

	logger = initializeLogger(config.Get())
}

func main() {
	audioGraph := graph.NewGraph(logger)

	videoReader, err := ytdlp.NewVideoReader(
		logger,
		ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile},
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

func initializeLogger(settings config.Config) *logging.Logger {
	logger := logging.NewLogger()
	logger.SetPanicOnError(true)

	for _, hCfg := range settings.Logging.Handlers {
		var level logging.Level
		switch hCfg.Level {
		case config.LoggingHandlerLevel_Debug:
			level = logging.LevelDebug
		case config.LoggingHandlerLevel_Info:
			level = logging.LevelInfo
		case config.LoggingHandlerLevel_Warn:
			level = logging.LevelWarn
		case config.LoggingHandlerLevel_Error:
			level = logging.LevelError
		case config.LoggingHandlerLevel_Fatal:
			level = logging.LevelFatal
		case config.LoggingHandlerLevel_Panic:
			level = logging.LevelPanic
		default:
			panic("invalid config.LoggingHandlerLevel")
		}

		var w io.Writer
		switch hCfg.Output {
		case "stdout":
			w = os.Stdout
		case "stderr":
			w = os.Stderr
		default:
			f, err := os.Open(hCfg.Output)

			if err != nil {
				fmt.Printf("failed to create logger: %s", err.Error())
				os.Exit(1)
			}

			w = f
		}

		switch hCfg.Type {
		case config.LoggingHandlerType_Pretty:
			logger.AddHandler(logging.NewPrettyHandler(w, level))
		case config.LoggingHandlerType_JSON:
			logger.AddHandler(logging.NewJsonHandler(w, level))
		default:
			panic("invalid config.LoggingHandlerLevel")
		}
	}

	return logger
}
