package discord

import (
	"accidentallycoded.com/fredboard/v3/discord/commands"
	"accidentallycoded.com/fredboard/v3/telemetry/logging"
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
	logger, err := bot.logger.NewChildLogger()
	if err != nil {
		bot.logger.FatalWithErr("failed to create onReady handler logger", err)
	}

	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("event", &event)

	logger.Info("Session opened")
}

func (bot *Bot) onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	logger, err := bot.logger.NewChildLogger()
	if err != nil {
		bot.logger.FatalWithErr("failed to create onInteractionCreate handler logger", err)
	}

	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("event", &event)

	logger.Debug("InterationCreate event received")

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
		onApplicationCommandInteraction(session, event.Interaction, logger)
	default:
		logger.Warn("ignoring interaction with unsupported / unknown interaction type")
	}
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger, err := log.NewChildLogger()
	if err != nil {
		logger.FatalWithErr("failed to create onApplicationCommandInteraction logger", err)
	}

	logger.SetData("session", session)
	logger.SetData("interaction", interaction)

	switch data := interaction.ApplicationCommandData(); data.Name {
	case "yt":
		go commands.YT(session, interaction, logger)
	case "fs":
		go commands.FS(session, interaction, logger)
	default:
		logger.Warn("Ignoring invalid command")
	}
}

func (bot *Bot) Start() {
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
	})

	if err != nil {
		bot.logger.FatalWithErr("failed to register new commands", err)
	}

	cmdLogger, err := bot.logger.NewChildLogger()
	if err != nil {
		bot.logger.PanicWithErr("failed to create command logger", err)
	}

	for _, cmd := range newCmds {
		cmdLogger.SetData("cmd", cmd)
		cmdLogger.Info("registered command")
	}

	cmdLogger.Close()

	err = session.Open()
	if err != nil {
		bot.logger.FatalWithErr("failed to open discord session", err)
	}

	bot.logger.Debug("opened discord session")

	defer func() {
		err := session.Close()
		if err != nil {
			bot.logger.ErrorWithErr("failed to close discord session", err)
			return
		}

		bot.logger.Info("closed discord session")
	}()
}
