package commands

import (
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord/voice"
	"accidentallycoded.com/fredboard/v3/sources"
	"accidentallycoded.com/fredboard/v3/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func YT(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
  logger, err := log.NewChildLogger()
  if err != nil {
    logger.FatalWithErr("failed to create logger for command.YT", err)
  }

  logger.SetData("session", &session)
  logger.SetData("interaction", &interaction)

	interactionData := interaction.ApplicationCommandData()

	session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	url, err := getRequiredApplicationCommandOption(interactionData, "url", discordgo.ApplicationCommandOptionString)
	if err != nil {
    logger.ErrorWithErr("failed to get required application option \"url\"", err)
		// TODO: Notify the user that there was an error
		return
	}

  logger.SetData("option.url", &url)
  logger.Debug("got application option \"url\"")

	encoder, err := codecs.NewOpusEncoder(48000, 2)
	if err != nil {
    logger.ErrorWithErr("failed to create opus encoder", err)
		// TODO: Notify the user that there was an error
		return
	}

  logger.SetData("encoder", &encoder)
  logger.Debug("created encoder")

	source, err := sources.NewYouTubeSource(url.StringValue(), sources.YOUTUBESTREAMQUALITY_BEST)
	if err != nil {
    logger.ErrorWithErr("failed to create YouTube source", err)
		// TODO: Notify the user that there was an error
		return
	}

  logger.SetData("source", &source)
  logger.Debug("set source")

	const mute = false
	const deaf = true
	voiceConnection, err := joinVoiceChannelIdOfInteractionCreator(session, interaction, mute, deaf)
	if err != nil {
    logger.ErrorWithErr("failed to join voice channel of interaction creator", err)
		// TODO: Notify the user that there was an error
		return
	}

  logger.SetData("voiceConnection", &voiceConnection)
  logger.Debug("joined voice channel of interaction creator")

	defer func() {
		err := voiceConnection.Disconnect()

		if err != nil {
      logger.ErrorWithErr("failed to close voice connection", err)
			// TODO: Notify the user that there was an error
			return
		}

    logger.Debug("closed voice connection")
	}()

	sink := voice.NewVoiceWriter(voiceConnection)

	time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

	encoder.EncodePCMS16LE(source, sink, 960)
}
