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
	logger.AddHandler(logging.NewPrettyHandler(os.Stdout, logging.LevelDebug))
	logger.SetPanicOnError(true)

	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load .env file", "error", err)
	}

	configLogger := logger.NewChildLogger()
	defer configLogger.Close()

	config.Init(configLogger)
	if ok, err := config.IsValid(); !ok {
		logger.Fatal("invalid config", "error", err)
	}

	logger.SetLevel(config.Config.Logging.Level)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

		web.Run(ctx, config.Config.Web.Address, childLogger)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		childLogger := logger.NewChildLogger()
		defer childLogger.Close()

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
