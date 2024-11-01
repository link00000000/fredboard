package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var logger = slog.Default()

func main() {
  if env, ok := os.LookupEnv("LOG_LEVEL"); ok {
    var level slog.Level
    switch strings.ToUpper(env) {
    case "WARN":
      level = slog.LevelWarn.Level()
    case "DEBUG":
      level = slog.LevelDebug.Level()
    case "INFO":
      level = slog.LevelInfo.Level()
    case "ERROR":
      level = slog.LevelInfo.Level()
    }

    slog.SetLogLoggerLevel(level)
    logger.Debug("Set log level", "level", level.String())
  }

  _, ok := os.LookupEnv("DISCORD_APP_ID")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_APP_ID")
    os.Exit(1)
  }

  _, ok = os.LookupEnv("DISCORD_PUBLIC_KEY")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_PUBLIC_KEY")
    os.Exit(1)
  }

  discordToken, ok := os.LookupEnv("DISCORD_TOKEN")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_TOKEN")
    os.Exit(1)
  }

  session, err := discordgo.New("Bot " + discordToken)
  if err != nil {
    logger.Error("Failed to create session", "error", err)
    os.Exit(1)
  }

  session.AddHandler(onReady)

  err = session.Open()
  if err != nil {
    logger.Error("Failed to open session", "error", err)
    os.Exit(1)
  }

  defer func () {
    if err := session.Close(); err != nil {
      logger.Error("Failed to close session gacefully", "error", err)
    } else {
      logger.Info("Closed session")
    }
  }()

  intSig := make(chan os.Signal, 1)
  signal.Notify(intSig, os.Interrupt)
  <-intSig
}

func onReady(s *discordgo.Session, e *discordgo.Ready) {
  logger.Info("Session opened", "event", e)
}
