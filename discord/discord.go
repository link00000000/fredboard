package discord

import (
	"accidentallycoded.com/fredboard/v3/config"
	"accidentallycoded.com/fredboard/v3/discord/commands"
	"accidentallycoded.com/fredboard/v3/telemetry"
	"github.com/bwmarrin/discordgo"
)

var logger = telemetry.GlobalLogger
var mainLtx *telemetry.Context

func Start(ltx *telemetry.Context) {
	mainLtx = ltx

	session, err := discordgo.New("Bot " + config.Config.Discord.Token)
	if err != nil {
		logger.Fatal("Failed to create discord session", err, ltx)
	}

	ltx.SetValue("session", session)
	logger.Debug("Created discord session", ltx)

	logger.Debug("Registering handlers", ltx)
	session.AddHandler(onReady)
	session.AddHandler(onInteractionCreate)

	logger.Debug("Registering commands", ltx)
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
		logger.Fatal("Failed to register new commands", err, ltx)
	}

	for _, cmd := range newCmds {
		lltx := logger.NewContext(ltx)
		lltx.SetValue("session", session)
		lltx.SetValue("cmd", cmd)

		logger.Info("Registered command", lltx)
	}

	err = session.Open()
	if err != nil {
		logger.Fatal("Failed to open discord session", err, ltx)
	}

	logger.Debug("Opened discord session", ltx)

	defer func() {
		err := session.Close()
		if err != nil {
			logger.Error("Failed to close discord session", err, ltx)
			return
		}

		logger.Info("Closed discord session", ltx)
	}()
}

func onReady(session *discordgo.Session, event *discordgo.Ready) {
	ltx := logger.NewContext(logger.RootCtx)
	defer ltx.Close()

	ltx.SetValue("session", session)
	ltx.SetValue("event", event)

	logger.Info("Session opened", ltx)
}

func onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	ltx := logger.NewContext(logger.RootCtx)
	defer ltx.Close()

	ltx.SetValue("session", session)
	ltx.SetValue("event", event)

	logger.Debug("InterationCreate event received", ltx)

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
		onApplicationCommandInteraction(session, event.Interaction)
	default:
		logger.Warn("Ignoring interaction with unsupported / unknown interaction type", ltx)
	}
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction) {
	ltx := logger.NewContext(logger.RootCtx)
	defer ltx.Close()

	ltx.SetValue("session", session)
	ltx.SetValue("interaction", interaction)

	switch data := interaction.ApplicationCommandData(); data.Name {
	case "yt":
		go commands.YT(session, interaction, ltx)
	case "fs":
		go commands.FS(session, interaction, ltx)
	default:
		logger.Warn("Ignoring invalid command", ltx)
	}
}
