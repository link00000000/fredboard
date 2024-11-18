package main

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"accidentallycoded.com/fredboard/v3/commands"
	"accidentallycoded.com/fredboard/v3/config"
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
					&discordgo.MessageEmbed{
						Type:        discordgo.EmbedTypeRich,
						Title:       "Error",
						Description: err.Error(),
						Color:       15548997, // Discord red
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
	case "fs":
		return commands.FS(session, interaction)
	default:
		return nil, ErrUnknownCommand
	}
}

func main() {
	config.Init()
	if ok, err := config.IsValid(); !ok {
		unwrappedErrs, ok := err.(interface{ Unwrap() []error })

		var errs []error
		if ok {
			errs = unwrappedErrs.Unwrap()
		} else {
			errs = []error{err}
		}

		logger.Error("Invalid config", "errors", errs)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(config.Config.Logging.Level)
	logger.Debug("Set log level", "level", config.Config.Logging.Level.String())

	session, err := discordgo.New("Bot " + config.Config.Discord.Token)
	if err != nil {
		logger.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	logger.Debug("Registering handlers")
	session.AddHandler(onReady)
	session.AddHandler(onInteractionCreate)

	logger.Debug("Registering commands")
	newCmds, err := session.ApplicationCommandBulkOverwrite(config.Config.Discord.AppId, "", []*discordgo.ApplicationCommand{
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
			Name:        "fs",
			Description: "Play a file from the file system",
			Options: []*discordgo.ApplicationCommandOption{
				&discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "path",
					Description: "Path to file on filesystem to play",
					Required:    true,
				},
				&discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "encoding",
					Description: "Encoding of the audio file. Either DCA0 or PCMS16LE",
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
		err := session.Close()
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
