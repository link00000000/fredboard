package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/discord"
	"accidentallycoded.com/fredboard/v3/internal/telemetry"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	_ "accidentallycoded.com/fredboard/v3/internal/telemetry/pprof"
	"accidentallycoded.com/fredboard/v3/internal/version"
)

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

	logger = logging.NewLogger()
	logger.SetPanicOnError(true)

	for _, hCfg := range config.Get().Logging.Handlers {
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

		_ = w
		_ = level
		switch hCfg.Type {
		case config.LoggingHandlerType_Pretty:
			//logger.AddHandler(logging.NewPrettyHandler(w, level))
		case config.LoggingHandlerType_JSON:
			//logger.AddHandler(logging.NewJsonHandler(w, level))
		default:
			panic("invalid config.LoggingHandlerLevel")
		}
	}
}

func main() {
	logger.Info("starting FredBoard", "version", version.String())

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	otelShutdown, err := telemetry.SetupOTelSDK("accidentallycoded.com/fredboard/v3/cmd/fredboard_server", ctx)
	if err != nil {
		// TODO: Properly handle
		fmt.Printf("fatal error during otel setup: %s", err.Error())
		os.Exit(1)
	}

	defer func() {
		err := errors.Join(err, otelShutdown(context.Background()))
		if err != nil {
			// TODO: Properly handle
			fmt.Printf("fatal error during otel shutdown: %s", err.Error())
			os.Exit(1)
		}
	}()

	bot := discord.NewBot(config.Get().Discord.AppId, config.Get().Discord.Token, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.Run(ctx)
	}()

	logger.Info("press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig

	logger.Info("received interrupt signal")

	cancel()
	wg.Wait()
}
