package commands

import (
	"errors"
	"strings"
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord"
	"accidentallycoded.com/fredboard/v3/sources"
	"github.com/bwmarrin/discordgo"
)

func FS(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.InteractionResponse, error) {
	interactionData := interaction.ApplicationCommandData()

	encoding, err := getRequiredApplicationCommandOption(interactionData, "encoding", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.Error("FS: Failed to get required application option", "session", session, "interaction", interaction, "option", "encoding", "error", err)
		return nil, err
	}

	path, err := getRequiredApplicationCommandOption(interactionData, "path", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.Error("FS: Failed to get required application option", "session", session, "interaction", interaction, "option", "path", "error", err)
		return nil, err
	}

	encoder, err := codecs.NewOpusEncoder(48000, 2)
	if err != nil {
		logger.Error("FS: Failed to create opus encoder", "session", session, "interaction", interaction, "error", err)
		return nil, err
	}

	source, err := sources.NewFSSource(path.StringValue())
	if err != nil {
		logger.Error("FS: Failed to create FS source", "session", session, "interaction", interaction, "error", err)
		return nil, err
	}

	const mute = false
	const deaf = true
	voiceConnection, err := joinVoiceChannelIdOfInteractionCreator(session, interaction, mute, deaf)
	if err != nil {
		logger.Error("FS: Failed to join voice channel of interaction creator", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return nil, err
	}

	logger.Debug("FS: Joined voice channel of interaction creator", "session", session, "interaction", interaction)

	defer func() {
		err := voiceConnection.Disconnect()

		if err != nil {
			logger.Error("FS: Failed to close voice connection", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			return
		}

		logger.Debug("FS: Closed voice channel", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	sink := discord.NewVoiceWriter(voiceConnection)

	time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

	switch strings.ToUpper(encoding.StringValue()) {
	case "PCMS16LE":
		encoder.EncodePCMS16LE(source, sink, 960)
	case "DCA0":
		encoder.EncodeDCA0(source, sink)
	default:
		logger.Error("FS: Unknown encoding", "encoding", encoding)
		return nil, errors.New("unknown encoding")
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	return response, nil
}