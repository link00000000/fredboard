package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
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
	// ytdlp videoReader1
	videoReader1, err := ytdlp.NewVideoReader(
		logger,
		ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile},
		"https://www.youtube.com/watch?v=F1oKhsy8wGw",
		ytdlp.YtdlpAudioQuality_BestAudio,
	)

	if err != nil {
		logger.Panic("failed to create video reader", "error", err)
	}

	// ffmpeg transcoder1
	transcoder1, err := ffmpeg.NewTranscoder(
		logger,
		ffmpeg.Config{ExePath: config.Get().Ffmpeg.ExePath},
		videoReader1,
		ffmpeg.Format_PCMSigned16BitLittleEndian,
		config.Get().Audio.SampleRateHz,
		config.Get().Audio.NumChannels,
	)

	if err != nil {
		logger.Panic("failed to create ffmpeg transcoder", "error", err)
	}

	defer transcoder1.Close()

	// ytdlp videoReader2
	videoReader2, err := ytdlp.NewVideoReader(
		logger,
		ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile},
		"https://www.youtube.com/watch?v=6f_yfQgV1w8",
		ytdlp.YtdlpAudioQuality_BestAudio,
	)

	if err != nil {
		logger.Panic("failed to create video reader", "error", err)
	}

	// ffmpeg transcoder2
	transcoder2, err := ffmpeg.NewTranscoder(
		logger,
		ffmpeg.Config{ExePath: config.Get().Ffmpeg.ExePath},
		videoReader2,
		ffmpeg.Format_PCMSigned16BitLittleEndian,
		config.Get().Audio.SampleRateHz,
		config.Get().Audio.NumChannels,
	)

	if err != nil {
		logger.Panic("failed to create ffmpeg transcoder", "error", err)
	}

	defer transcoder2.Close()

	// file WriterNode
	outputFile, err := os.Create("output.wav")

	if err != nil {
		logger.Panic("failed to create output file")
	}

	defer outputFile.Close()

	readerNode1 := graph.NewReaderNode(logger, transcoder1, 0x8000)
	readerNode2 := graph.NewReaderNode(logger, transcoder2, 0x8000)
	mixerNode := graph.NewMixerNode(logger)
	writerNode := graph.NewWriterNode(logger, outputFile)

	audioGraph := graph.NewGraph(logger)
	audioGraph.AddNode(readerNode1)
	audioGraph.AddNode(readerNode2)
	audioGraph.AddNode(mixerNode)
	audioGraph.AddNode(writerNode)
	audioGraph.CreateConnection(readerNode1, mixerNode)
	audioGraph.CreateConnection(readerNode2, mixerNode)
	audioGraph.CreateConnection(mixerNode, writerNode)

	logger.Info("starting audio graph")

	reader1Done, reader2Done := false, false
	for {
		audioGraph.Tick()

		if err := audioGraph.Err(); err != nil {
			if errors.Is(readerNode1.Err(), io.EOF) {
				audioGraph.RemoveConnection(readerNode1, mixerNode)
				audioGraph.RemoveNode(readerNode1)
				reader1Done = true
			}

			if errors.Is(readerNode2.Err(), io.EOF) {
				audioGraph.RemoveConnection(readerNode2, mixerNode)
				audioGraph.RemoveNode(readerNode2)
				reader2Done = true
			}

			if !errors.Is(err, io.EOF) {
				logger.Panic("failed to tick audio graph", "error", err)
			}
		}

		if reader1Done && reader2Done {
			break
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
