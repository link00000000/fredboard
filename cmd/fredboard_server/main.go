package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/discord"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"accidentallycoded.com/fredboard/v3/internal/web"
	"github.com/joho/godotenv"
)

// These values are populated by the linker using -ldflags "-X main.version=x.x.x -X main.commit=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
var (
	buildVersion string
	buildCommit  string
)

func main() {
	var logger = logging.NewLogger()
	logger.AddHandler(logging.NewPrettyHandler(os.Stdout))
	logger.SetPanicOnError(true)

	err := godotenv.Load()
	if err != nil {
		logger.ErrorWithErr("failed to load .env file", err)
	}

	configLogger := logger.NewChildLogger()
	defer configLogger.Close()

	config.Init(configLogger)
	if ok, err := config.IsValid(); !ok {
		logger.FatalWithErr("invalid config", err)
	}

	logger.SetLevel(config.Config.Logging.Level)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	go func() {
		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

		wg.Add(1)
		defer wg.Done()

		web.Run(ctx, config.Config.Web.Address, childLogger)
	}()

	go func() {
		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

		wg.Add(1)
		defer wg.Done()

		bot := discord.NewBot(config.Config.Discord.AppId, config.Config.Discord.Token, childLogger)
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
