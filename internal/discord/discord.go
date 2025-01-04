package discord

import (
	"context"

	"accidentallycoded.com/fredboard/v3/internal/discord/commands"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	logger *logging.Logger
	appId  string
	token  string
}

func NewBot(appId, token string, logger *logging.Logger) Bot {
	return Bot{appId: appId, token: token, logger: logger}
}

func (bot *Bot) onReady(session *discordgo.Session, event *discordgo.Ready) {
	logger := bot.logger.NewChildLogger()

	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("event", &event)

	logger.Info("session opened")
}

func (bot *Bot) onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	logger := bot.logger.NewChildLogger()

	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("event", &event)

	logger.Debug("event received")

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
		onApplicationCommandInteraction(session, event.Interaction, logger)
	default:
		logger.Warn("ignoring interaction with unsupported / unknown interaction type")
	}
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

	switch data := interaction.ApplicationCommandData(); data.Name {
	case "yt":
		go commands.YT(session, interaction, logger)
	case "fs":
		go commands.FS(session, interaction, logger)
	case "leave":
		go commands.Leave(session, interaction, logger)
	default:
		logger.Warn("ignoring invalid command")
	}
}

func (bot *Bot) Run(ctx context.Context) {
	session, err := discordgo.New("Bot " + bot.token)
	if err != nil {
		bot.logger.FatalWithErr("failed to create discord session", err)
	}

	bot.logger.SetData("session", &session)
	bot.logger.Debug("created discord session")

	bot.logger.Debug("registering handlers")
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onInteractionCreate)

	bot.logger.Debug("registering commands")
	newCmds, err := session.ApplicationCommandBulkOverwrite(bot.appId, "", []*discordgo.ApplicationCommand{
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
		&discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "leave",
			Description: "Stop playing and leave the voice channel",
		},
	})

	if err != nil {
		bot.logger.FatalWithErr("failed to register new commands", err)
	}

	cmdLogger := bot.logger.NewChildLogger()

	for _, cmd := range newCmds {
		cmdLogger.SetData("cmd", &cmd)
		cmdLogger.Info("registered command")
	}

	cmdLogger.Close()

	err = session.Open()
	if err != nil {
		bot.logger.FatalWithErr("failed to open discord session", err)
	}

	defer bot.logger.Info("discord bot shutdown")

	defer func() {
		err := session.Close()
		if err != nil {
			bot.logger.ErrorWithErr("failed to close discord session", err)
			return
		}

		bot.logger.Info("discord session closed")
	}()

	<-ctx.Done()
	bot.logger.Info("stopping discord bot")
}
