package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"accidentallycoded.com/fredboard/v3/sources"
	"github.com/bwmarrin/discordgo"
)

var logger = slog.Default()

type DiscordBot struct {
	Session *discordgo.Session
}

func NewDiscordBot(appId, pubilcKey, token string) (*DiscordBot, error) {
  b := &DiscordBot{}

	if session, err := discordgo.New("Bot " + token); err != nil {
    return nil, err
  } else {
    b.Session = session
  }

	logger.Debug("Registering handlers")
	b.Session.AddHandler(b.onReady)
	b.Session.AddHandler(b.onInteractionCreate)

	logger.Debug("Registering commands")
	newCmds, err := b.Session.ApplicationCommandBulkOverwrite(appId, "", []*discordgo.ApplicationCommand{
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
	})

	if err != nil {
		return nil, err
	}

	for _, cmd := range newCmds {
		logger.Info("Registered new command", "name", cmd.Name, "type", cmd.Type)
	}

	b.Session.Open()
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Implements [io.Closer]
func (b *DiscordBot) Close() error {
	return b.Session.Close()
}

func (b *DiscordBot) onReady(s *discordgo.Session, e *discordgo.Ready) {
	logger.Info("Session opened", "event", e)
}

func (b *DiscordBot) onInteractionCreate(s *discordgo.Session, e *discordgo.InteractionCreate) {
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

      b.ytCommand(url, e)

			return
		}

		logger.Warn("Command interaction unknown", "event", e)
	default:
		logger.Warn("Interaction type not supported", "event", e)
	}
}

func (b *DiscordBot) ytCommand(url string, event *discordgo.InteractionCreate) {
  s := b.Session

  res := &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: fmt.Sprintf("YouTube video will now play... { %#+v }", url),
    },
  }

  err := s.InteractionRespond(event.Interaction, res)
  if err != nil {
    logger.Error("Failed to respond to interaction /yt", "event", event, "error", err)
    return
  }
  logger.Debug("Responded to interaction /yt", "event", event, "response", res)

  g, err := s.State.Guild(event.GuildID)
  if err != nil {
    logger.Error("Failed to get guild", "event", event, "guildId", event.GuildID)
    return
  }
  logger.Debug("Got guild for interaction", "event", event, "guild", g)

  var vcId string
  for _, vs := range g.VoiceStates {
    if vs.UserID == event.Member.User.ID {
      vcId = vs.ChannelID
      break
    }
  }

  if vcId == "" {
    logger.Error("Sender is not in an accessible voice channel", "event", event)
    return
  }
  logger.Debug("Sender in voice channel", "event", event, "voiceChannelId", vcId)

  vc, err := s.ChannelVoiceJoin(event.GuildID, vcId, false, true)
  if err != nil {
    logger.Error("Failed to join voice channel", "event", event, "voiceChannelId", vcId, "error", err)
    return
  }
  defer vc.Disconnect()

  time.Sleep(250 * time.Millisecond)

  file, err := os.Open("assets/airhorn.dca")
  if err != nil {
    logger.Error("Failed to open assets/airhorn.dca", "error", err)
    return
  }
  defer func() {
    err := file.Close()
    if err != nil {
      logger.Error("Failed to close assets/airhorn.dca", "error", err)
    }
  }()

  err = vc.Speaking(true)
  if err != nil {
    logger.Error("Error while setting speaking status: true", "event", event, "voiceCHannel", vc)
    return
  }

  defer func() {
    if err := vc.Speaking(false); err != nil {
      logger.Error("Error while setting speaking status: false", "event", event, "voiceCHannel", vc)
    }
  }()

  yt := sources.NewYouTubeStream(url, sources.YOUTUBEQUALITY_WORST)

  dataChan := make(chan []byte, 32)
  errChan := make(chan error, 32)
  done := make(chan bool, 1)

  go func() {
    for {
      select {
      case data := <-dataChan:
        vc.OpusSend <- data
      case err := <-errChan:
        logger.Error("Error from YouTubeStream", "error", err, "yt", yt)
      case <-done:
        break
      }
    }
  }()

  if err = yt.Start(dataChan, errChan); err != nil {
    logger.Error("Failed to start YouTubeStream", "error", err, "yt", yt)
    return
  }

  defer func() {
    if err := yt.Wait(); err != nil {
      logger.Error("Failed to wait for YouTubeStream", "error", err, "yt", yt)
    } else {
      done <- true
    }
  }()

  return
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

  bot, err := NewDiscordBot(discordAppId, discordPublicKey, discordToken)
  if err != nil {
    logger.Error("Failed to create discord bot", "error", err)
  }

	defer func() {
		if err := bot.Close(); err != nil {
			logger.Error("Failed to close bot gacefully", "error", err)
		} else {
			logger.Info("Closed bot")
		}
	}()

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)
	<-intSig
}

