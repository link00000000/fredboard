package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/discord"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"accidentallycoded.com/fredboard/v3/internal/web"
)

// These values are populated by the linker using -ldflags "-X main.version=x.x.x -X main.commit=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
var (
	buildVersion string
	buildCommit  string
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Printf("failed to initialize config: %s", err.Error())
		os.Exit(1)
	}

	if ok, errs := config.Validate(); !ok {
		fmt.Printf("invalid config:\n%s", errors.Join(errs...))
		os.Exit(1)
	}

	var logger = logging.NewLogger()
	logger.SetPanicOnError(true)

	settings := config.Get()
	for _, handlerConfig := range settings.Logging.Handlers {
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

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

		web.Run(ctx, *settings.Web.Address, childLogger)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

		bot := discord.NewBot(*settings.Discord.AppId, *settings.Discord.Token, childLogger)
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
