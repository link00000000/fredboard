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

	configLogger := logger.NewChildLogger()
	defer configLogger.Close()

	config.Init(configLogger)
	if ok, err := config.IsValid(); !ok {
		logger.FatalWithErr("Invalid config", err)
	}

	logger.SetLevel(config.Config.Logging.Level)

	go func() {
		childLogger := logger.NewChildLogger()

		defer childLogger.Close()

		web.Start(config.Config.Web.Address, childLogger)
	}()

	go func() {
		childLogger := logger.NewChildLogger()

		defer childLogger.Close()

		bot := discord.NewBot(config.Config.Discord.AppId, config.Config.Discord.Token, childLogger)
		bot.Start()
	}()

	logger.Info("press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}
