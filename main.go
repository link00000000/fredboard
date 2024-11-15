package main

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"accidentallycoded.com/fredboard/v3/commands"
	"github.com/bwmarrin/discordgo"
)

var logger = slog.Default()

var ErrUnknownCommand = errors.New("unknown command")

func onReady(session *discordgo.Session, e *discordgo.Ready) {
	logger.Info("Session opened", "event", e)
}

func onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	logger.Debug("InteractionCreate event received", "guildId", event.GuildID, "channelId", event.ChannelID)

  var err error
  var response *discordgo.InteractionResponse

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
    response, err = onApplicationCommandInteraction(session, event.Interaction)
  default:
    err = errors.New("unsupported interaction type")
  }

  if err != nil {
    logger.Error("onInteractionCreate: Error while handling interaction", "session", session, "event", event, "error", err)

    err := session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
      Type: discordgo.InteractionResponseChannelMessageWithSource,
      Data: &discordgo.InteractionResponseData{
        Content: "There was an error while handling interaction",
        Embeds: []*discordgo.MessageEmbed{
          &discordgo.MessageEmbed {
            Type: discordgo.EmbedTypeRich,
            Title: "Error",
            Description: err.Error(),
            Color: 15548997, // Discord red
          },
        },
      },
    })

    if err != nil {
      logger.Error("onInteractionCreate: Error while responding to interaction", "session", session, "event", event, "error", err)
    }

    return
  }

  if response != nil {
    err := session.InteractionRespond(event.Interaction, response)

    if err != nil {
      logger.Error("onInteractionCreate: Error while responding to interaction", "session", session, "event", event, "error", err)
    }
  }
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.InteractionResponse, error) {
  data := interaction.ApplicationCommandData()

  switch data.Name {
  case "yt":
    return commands.Yt(session, interaction)
  case "dca":
    return commands.Dca(session, interaction)
  case "ogg":
    return commands.Ogg(session, interaction)
  default:
    return nil, ErrUnknownCommand
  }
}

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

	session, err := discordgo.New("Bot " + discordToken);
  if err != nil {
		logger.Error("Failed to create bot", "error", err)
    os.Exit(1)
	}

	logger.Debug("Registering handlers")
	session.AddHandler(onReady)
	session.AddHandler(onInteractionCreate)

	logger.Debug("Registering commands")
	newCmds, err := session.ApplicationCommandBulkOverwrite(discordAppId, "", []*discordgo.ApplicationCommand{
		&discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "yt",
			Description: "Play a YouTube video",
			Options: []*discordgo.ApplicationCommandOption{
				&discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "Url to the YouTube video to play",
					Required:    true,
				},
			},
		},
		&discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "ogg",
			Description: "Play an ogg file from the filesystem",
			Options: []*discordgo.ApplicationCommandOption{
				&discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "path",
					Description: "Path to file on filesystem to play",
					Required:    true,
				},
			},
		},
		&discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "dca",
			Description: "Play a dca file from the filesystem",
			Options: []*discordgo.ApplicationCommandOption{
				&discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "path",
					Description: "Path to file on filesystem to play",
					Required:    true,
				},
			},
		},
	})

	if err != nil {
		logger.Error("Failed to register new commands", "error", err)
    os.Exit(1)
	}

	for _, cmd := range newCmds {
		logger.Debug("Registered new command", "name", cmd.Name, "type", cmd.Type)
	}

  err = session.Open()
  if err != nil {
    logger.Error("Failed to open discord session", "error", err)
    os.Exit(1)
  } else {
    logger.Debug("Opened discord session", "session", session)
  }

	defer func() {
		err := session.Close();
    if err != nil {
			logger.Error("Failed to close discord session", "error", err)
		} else {
			logger.Info("Closed discord session")
		}
	}()

  logger.Info("Press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}
