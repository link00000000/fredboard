package commands

import (
	"strings"
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord/voice"
	"accidentallycoded.com/fredboard/v3/sources"
	"accidentallycoded.com/fredboard/v3/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func FS(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

	interactionData := interaction.ApplicationCommandData()

	session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	encoding, err := getRequiredApplicationCommandOption(interactionData, "encoding", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.ErrorWithErr("failed to get required application option \"encoding\"", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.SetData("option.encoding", &encoding)
	logger.Debug("got application option \"encoding\"")

	path, err := getRequiredApplicationCommandOption(interactionData, "path", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.ErrorWithErr("failed to get required application option \"path\"", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.SetData("option.path", &path)
	logger.Debug("got application option \"path\"")

	encoder, err := codecs.NewOpusEncoder(48000, 2)
	if err != nil {
		logger.ErrorWithErr("failed to create opus encoder", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.SetData("encoder", &encoder)
	logger.Debug("created encoder")

	source := sources.NewFSSource(path.StringValue())

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

	err = source.Start()
	if err != nil {
		logger.ErrorWithErr("failed to start source", err)
		return
	}

	switch strings.ToUpper(encoding.StringValue()) {
	case "PCMS16LE":
		go encoder.EncodePCMS16LE(source, sink, 960)
	case "DCA0":
		go encoder.EncodeDCA0(source, sink)
	default:
		logger.Error("unknown encoding")
		// TODO: Notify the user that there was an error
	}

	err = source.Wait()
	if err != nil {
		logger.ErrorWithErr("error while waiting for source", err)
		return
	}
}
