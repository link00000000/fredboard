package main

import (
	"os"
	"os/signal"

	"accidentallycoded.com/fredboard/v3/config"
	"accidentallycoded.com/fredboard/v3/discord"
	"accidentallycoded.com/fredboard/v3/telemetry"
	"accidentallycoded.com/fredboard/v3/web"
)

func main() {
  var logger = telemetry.NewLogger([]telemetry.Handler{})
	//logger.AddHandler(telemetry.NewJsonHandler(os.Stdout))
	//logger.AddHandler(telemetry.NewPrettyHandler(os.Stdout))

	var ltx = logger.RootCtx

  


	config.Init()
	if ok, err := config.IsValid(); !ok {
		logger.Fatal("Invalid config", err, ltx)
	}

	logger.RootCtx.SetValue("config", config.Config)
	logger.SetLevel(config.Config.Logging.Level)

	logger.Debug("Loaded config", logger.RootCtx)

	go func() {
		webLtx := logger.NewContext(ltx)
		defer webLtx.Close()

		web.Start(webLtx)
	}()

	go func() {
		discordLtx := logger.NewContext(ltx)
		defer discordLtx.Close()

		discord.Start(discordLtx)
	}()

	logger.Info("Press ^c to exit", ltx)

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}
