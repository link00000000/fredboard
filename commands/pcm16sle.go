package commands

import (
	"fmt"
	"os"
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord"
	"github.com/bwmarrin/discordgo"
)

func PCMS16LE(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.InteractionResponse, error) {
	interactionData := interaction.ApplicationCommandData()

	path, err := getRequiredApplicationCommandOption(interactionData, "path", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.Error("PCMS16LE: Failed to get required application option", "session", session, "interaction", interaction, "name", "path", "error", err)
		return nil, err
	}

	go playPCMS16LEFile(session, interaction, path.StringValue())

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}

	return response, nil
}

func playPCMS16LEFile(session *discordgo.Session, interaction *discordgo.Interaction, path string) {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to open file", "path", path, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playPCMS16LEFile: Opened file", "path", path, "file", f)

	defer func() {
		err := f.Close()

		if err != nil {
			logger.Error("playPCMS16LEFile: Failed to close file", "file", f)
			return
		}

		logger.Debug("playPCMS16LEFile: Closed file", "file", f)
	}()

	voiceChannelId, err := findVoiceChannelIdOfInteractionCreator(session, interaction)
	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to find join voice channel of interaction creator", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playPCMS16LEFile: Found voice channel of interaction creator", "session", session, "interaction", interaction, "voiceChannelId", voiceChannelId)

	const mute = false
	const deaf = true
	voiceConnection, err := session.ChannelVoiceJoin(interaction.GuildID, voiceChannelId, mute, deaf)

	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to join voice channel", "session", session, "interaction", interaction, "voiceChannelId", voiceChannelId)
		// TODO: Notify the user that there was an error
		return
	}

	defer func() {
		err := voiceConnection.Disconnect()

		if err != nil {
			logger.Error("playPCMS16LEFile: Failed to close voice connection", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			return
		}

		logger.Debug("playPCMS16LEFile: Closed voice channel", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

	session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Playing %s", f.Name),
		},
	})

	// TODO: Move constants to command options
	encoder, err := codecs.NewOpusEncoder(48000, 2)

	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to create opus encoder", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "error", err)
	}

	logger.Debug("playPCMS16LEFile: Created opus encoder", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)

	voiceWriter := discord.NewVoiceWriter(voiceConnection)
	logger.Debug("playPCMS16LEFile: Created voice writer", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "voiceWriter", voiceWriter)

	err = voiceConnection.Speaking(true)
	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playPCMS16LEFile: Set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)

	defer func() {
		err = voiceConnection.Speaking(false)

		if err != nil {
			logger.Error("playPCMS16LEFile: Failed to set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			// TODO: Notify the user that there was an error
			return
		}

		logger.Debug("playPCMS16LEFile: Set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	// TODO: Move constant to command option
	err = encoder.EncodePCMS16LE(f, voiceWriter, 960)

	if err != nil {
		logger.Error("playPCMS16LEFile: Failed to encode to voice writer", "session", session, "interaction", interaction, "voiceConnection", voiceConnection, "voiceWriter", voiceWriter, "file", f, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("playPCMS16LEFile: Finished encoding", "session", session, "interaction", interaction, "file", f)
}
