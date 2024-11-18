package commands

import (
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord"
	"accidentallycoded.com/fredboard/v3/sources"
	"github.com/bwmarrin/discordgo"
)

func YT(session *discordgo.Session, interaction *discordgo.Interaction) {
	interactionData := interaction.ApplicationCommandData()

	session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	url, err := getRequiredApplicationCommandOption(interactionData, "url", discordgo.ApplicationCommandOptionString)
	if err != nil {
		logger.Error("YT: Failed to get required application option", "session", session, "interaction", interaction, "option", "url", "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	encoder, err := codecs.NewOpusEncoder(48000, 2)
	if err != nil {
		logger.Error("YT: Failed to create opus encoder", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	source, err := sources.NewYouTubeSource(url.StringValue(), sources.YOUTUBESTREAMQUALITY_BEST)
	if err != nil {
		logger.Error("YT: Failed to create YouTube source", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	const mute = false
	const deaf = true
	voiceConnection, err := joinVoiceChannelIdOfInteractionCreator(session, interaction, mute, deaf)
	if err != nil {
		logger.Error("YT: Failed to join voice channel of interaction creator", "session", session, "interaction", interaction, "error", err)
		// TODO: Notify the user that there was an error
		return
	}

	logger.Debug("YT: Joined voice channel of interaction creator", "session", session, "interaction", interaction)

	defer func() {
		err := voiceConnection.Disconnect()

		if err != nil {
			logger.Error("YT: Failed to close voice connection", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
			// TODO: Notify the user that there was an error
			return
		}

		logger.Debug("YT: Closed voice channel", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
	}()

	sink := discord.NewVoiceWriter(voiceConnection)

	time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

	encoder.EncodePCMS16LE(source, sink, 960)
}
