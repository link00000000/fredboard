package commands

import (
	"fmt"
	"os"
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord"
	"github.com/bwmarrin/discordgo"
)

const (
	sampleRate  = 48000 // Hz
	numChannels = 2
)

func DCA0(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.InteractionResponse, error) {
	interactionData := interaction.ApplicationCommandData()

	path, err := getRequiredApplicationCommandOption(interactionData, "path", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.Error("DCA: Failed to get required application option", "session", session, "interaction", interaction, "name", "path", "error", err)
		return nil, err
	}

	go playDCA0File(session, interaction, path.StringValue())

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	return response, nil
}

func playDCA0File(session *discordgo.Session, interaction *discordgo.Interaction, path string) {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("playDCA0File: Failed to open file", "path", path, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playDCA0File: Opened file", "path", path, "file", f)

	defer func() {
		err := f.Close()

		if err != nil {
			logger.Error("playDCA0File: Failed to close file", "file", f)
			return
		}

		logger.Debug("playDCA0File: Closed file", "file", f)
	}()

	const mute = false
	const deaf = true
	voiceConnection, err := joinVoiceChannelIdOfInteractionCreator(session, interaction, mute, deaf)
	if err != nil {
		logger.Error("playDCA0File: Failed to join voice channel of interaction creator", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playDCA0File: Joined voice channel of interaction creator", "session", session, "interaction", interaction)

	defer func() {
		err := voiceConnection.Disconnect()

		if err != nil {
			logger.Error("playDCA0File: Failed to close voice connection", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			return
		}

		logger.Debug("playDCA0File: Closed voice channel", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

	session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Playing %s", f.Name),
		},
	})

	encoder, err := codecs.NewOpusEncoder(sampleRate, numChannels)

	if err != nil {
		logger.Error("playDCA0File: Failed to create opus encoder", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "error", err)
	}

	logger.Debug("playDCA0File: Created opus encoder", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)

	voiceWriter := discord.NewVoiceWriter(voiceConnection)
	logger.Debug("playPCMS16LEFile: Created voice writer", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "voiceWriter", voiceWriter)

	err = voiceConnection.Speaking(true)
	if err != nil {
		logger.Error("playDCA0File: Failed to set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playDCA0File: Set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)

	defer func() {
		err = voiceConnection.Speaking(false)

		if err != nil {
			logger.Error("playDCA0File: Failed to set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			// TODO: Notify the user that there was an error
			return
		}

		logger.Debug("playDCA0File: Set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	err = encoder.EncodeDCA0(f, voiceWriter)

	if err != nil {
		logger.Error("playDCA0File: Failed to encode to voice writer", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "voiceWriter", voiceWriter, "file", f, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playDCA0File: Finished encoding", "session", session, "interaction", interaction, "file", f)
}
