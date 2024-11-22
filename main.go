package main

import (
	"os"
	"os/signal"

	"accidentallycoded.com/fredboard/v3/commands"
	"accidentallycoded.com/fredboard/v3/config"
	"accidentallycoded.com/fredboard/v3/telemetry"
	"accidentallycoded.com/fredboard/v3/web"
	"github.com/bwmarrin/discordgo"
)

var logger = telemetry.NewLogger([]telemetry.Handler{
	telemetry.NewPrettyHandler(os.Stdout),
})

func onReady(session *discordgo.Session, e *discordgo.Ready) {
	logger.InfoWithContext("Session opened", telemetry.Context{"session": session, "event": e})
}

func onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	logger.DebugWithContext("InterationCreate event received", telemetry.Context{"session": session, "event": event})

	switch event.Data.Type() {
	case discordgo.InteractionApplicationCommand:
		onApplicationCommandInteraction(session, event.Interaction)
	default:
		logger.WarnWithContext("Ignoring interaction with unsupported / unknown interaction type", telemetry.Context{"session": session, "event": event})
	}
}

func onApplicationCommandInteraction(session *discordgo.Session, interaction *discordgo.Interaction) {
	switch data := interaction.ApplicationCommandData(); data.Name {
	case "yt":
		go commands.YT(session, interaction)
	case "fs":
		go commands.FS(session, interaction)
	default:
		logger.WarnWithContext("Ignoring invalid command", telemetry.Context{"session": session, "interaction": interaction, "commandData": data})
	}
}

func main() {
	go web.Start()

	config.Init()
	if ok, err := config.IsValid(); !ok {
		logger.Fatal("Invalid config", err)
	} else {
		logger.DebugWithContext("Loaded config", telemetry.Context{"config": config.Config})
	}

	logger.SetLevel(config.Config.Logging.Level)
	logger.DebugWithContext("Set log level", telemetry.Context{"level": config.Config.Logging.Level})

	session, err := discordgo.New("Bot " + config.Config.Discord.Token)
	if err != nil {
		logger.Fatal("Failed to create discord session", err)
	} else {
		logger.DebugWithContext("Created discord session", telemetry.Context{"session": session})
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
		logger.FatalWithContext("Failed to register new commands", err, telemetry.Context{"session": session})
	}

	for _, cmd := range newCmds {
		logger.InfoWithContext("Registered command", telemetry.Context{"cmd": cmd})
	}

	err = session.Open()
	if err != nil {
		logger.FatalWithContext("Failed to open discord session", err, telemetry.Context{"session": session})
	} else {
		logger.DebugWithContext("Opened discord session", telemetry.Context{"session": session})
	}

	defer func() {
		err := session.Close()
		if err != nil {
			logger.Error("Failed to close discord session", err)
		} else {
			logger.Info("Closed discord session")
		}
	}()

	logger.Info("Press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}
