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

	logger.Info("session opened", "session", session, "event", event)
}

func (bot *Bot) onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	logger := bot.logger.NewChildLogger()

	defer logger.Close()

	logger.Debug("event received", "session", session, "event", event)

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
		onApplicationCommandInteraction(session, event.Interaction, logger)
	default:
		logger.Warn("ignoring interaction with unsupported / unknown interaction type")
	}
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()

	// TODO: recover from any panics
	switch data := interaction.ApplicationCommandData(); data.Name {
	case "yt":
		go commands.Yt(logger, session, interaction)
	case "join":
		go commands.Join(logger, session, interaction)
	case "leave":
		go commands.Leave(logger, session, interaction)
	default:
		logger.Warn("ignoring invalid command", "session", session, "interaction", interaction, "data", data)
	}
}

func (bot *Bot) Run(ctx context.Context) {
	session, err := discordgo.New("Bot " + bot.token)
	if err != nil {
		bot.logger.Fatal("failed to create discord session", "error", err)
	}

	bot.logger.Debug("created discord session", "session", session)

	bot.logger.Debug("registering handlers", "session", session)
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onInteractionCreate)

	bot.logger.Debug("registering commands", "session", session)
	newCmds, err := session.ApplicationCommandBulkOverwrite(bot.appId, "", []*discordgo.ApplicationCommand{
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "yt",
			Description: "Play a YouTube video",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "Url to the YouTube video to play",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "join",
			Description: "Join the voice channel",
		},
		{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "leave",
			Description: "Leave the voice channel",
		},
	})

	if err != nil {
		bot.logger.Fatal("failed to register new commands", "session", session, "error", err)
	}

	for _, cmd := range newCmds {
		bot.logger.Info("registered command", "session", session, "cmd", cmd)
	}

	err = session.Open()
	if err != nil {
		bot.logger.Fatal("failed to open discord session", "session", session, "error", err)
	}

	defer bot.logger.Info("discord bot shutdown", "session", session)

	defer func() {
		err := session.Close()
		if err != nil {
			bot.logger.Fatal("failed to close discord session", "session", session, "error", err)
			return
		}

		bot.logger.Info("discord session closed", "session", session)
	}()

	<-ctx.Done()
	bot.logger.Info("stopping discord bot", "session", session)
}
