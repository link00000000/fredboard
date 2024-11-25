package main

import (
	"os"
	"os/signal"

	"accidentallycoded.com/fredboard/v3/config"
	"accidentallycoded.com/fredboard/v3/discord"
	"accidentallycoded.com/fredboard/v3/telemetry/logging"
	"accidentallycoded.com/fredboard/v3/web"
)

func main() {
	var logger = logging.NewLogger()
	logger.AddHandler(logging.NewPrettyHandler(os.Stdout))
	logger.SetPanicOnError(true)

	config.Init()
	if ok, err := config.IsValid(); !ok {
		logger.FatalWithErr("Invalid config", err)
	}

	logger.SetData("config", config.Config)
	logger.SetLevel(config.Config.Logging.Level)

	logger.Debug("Loaded config")

	go func() {
		childLogger, err := logger.NewChildLogger()
		if err != nil {
			logger.PanicWithErr("failed to create logger for web", err)
		}

		defer childLogger.Close()

		web := web.NewWeb(childLogger)
		web.Start()
	}()

	go func() {
		childLogger, err := logger.NewChildLogger()
		if err != nil {
			logger.PanicWithErr("failed to create logger for discord", err)
		}

		defer childLogger.Close()

		bot := discord.NewBot(config.Config.Discord.AppId, config.Config.Discord.Token, childLogger)
		bot.Start()
	}()

	logger.Info("Press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}
