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

  discordAppId, ok := os.LookupEnv("DISCORD_APP_ID")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_APP_ID")
    os.Exit(1)
  }
  logger.Debug("Read environment variable DISCORD_APP_ID", "value", discordAppId)

  discordPublicKey, ok := os.LookupEnv("DISCORD_PUBLIC_KEY")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_PUBLIC_KEY")
    os.Exit(1)
  }
  logger.Debug("Read environment variable DISCORD_PUBLIC_KEY", "value", discordPublicKey)

  discordToken, ok := os.LookupEnv("DISCORD_TOKEN")
  if !ok {
    logger.Error("Required environment variable not set: DISCORD_TOKEN")
    os.Exit(1)
  }
  logger.Debug("Read environment variable DISCORD_TOKEN", "value", "[secret]")

  session, err := discordgo.New("Bot " + discordToken)
  if err != nil {
    logger.Error("Failed to create session", "error", err)
    os.Exit(1)
  }

  logger.Debug("Registering handlers")
  session.AddHandler(onReady)
  session.AddHandler(onInteractionCreate)

  logger.Debug("Registering commands")
  newCmds, err := session.ApplicationCommandBulkOverwrite(discordAppId, "", []*discordgo.ApplicationCommand{
    &discordgo.ApplicationCommand{
      Type: discordgo.ChatApplicationCommand,
      Name: "yt",
      Description: "Play a YouTube video",
      Options: []*discordgo.ApplicationCommandOption{
        &discordgo.ApplicationCommandOption{
          Type: discordgo.ApplicationCommandOptionString,
          Name: "url",
          Description: "Url to the YouTube video to play",
          Required: true,
        },
      },
    },
  })
  if err != nil {
    logger.Error("Failed to register commands", "error", err)
    os.Exit(1)
  }

  for _, cmd := range newCmds {
    logger.Debug("Registered new command", "name", cmd.Name, "type", cmd.Type)
  }

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

func onInteractionCreate(s *discordgo.Session, e *discordgo.InteractionCreate) {
  logger.Debug("InteractionCreate event received", "guildId", e.GuildID, "channelId", e.ChannelID)

  switch data := e.Data.(type) {
  case discordgo.ApplicationCommandInteractionData:
    if data.Name == "yt" {
      logger.Debug("Interaction matched type /yt", "event", e)

      var url string
      for _, opt := range data.Options {
        switch opt.Name {
        case "url":
          if opt.Type != discordgo.ApplicationCommandOptionString {
            logger.Error("Option received does not match the registered type for interaction /yt",
              "name", "url",
              "registeredType", discordgo.ApplicationCommandOptionString,
              "receivedOption", opt,
            )
            continue
          }

          url = opt.StringValue()
          logger.Debug("Received option for /yt", "name", "url", "value", opt.StringValue())

        default:
          logger.Warn("Received unknown option for interaction /yt", "option", opt)
        }
      }

      if len(url) == 0 {
        logger.Error("Did not receive required option for interaction /yt", "name", "url", "event", e)
        return
      }

      res := &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Content: "YouTube video will now play...",
        },
      }

      err := s.InteractionRespond(e.Interaction, res)
      if err != nil {
        logger.Error("Failed to respond to interaction /yt", "event", e, "error", err)
        return
      }

      logger.Debug("Responded to interaction /yt", "event", e, "response", res)
      return
    }

    logger.Warn("Command interaction unknown", "event", e)
  default:
    logger.Warn("Interaction type not supported", "event", e)
  }
}
