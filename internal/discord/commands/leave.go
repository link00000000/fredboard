package commands

import (
	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/discord/voice"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Leave(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()
	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

	voiceConn, ok := session.VoiceConnections[interaction.GuildID]
	if !ok {
		logger.Info("command /leave executed without any active voice connections in guild")

		err := interactions.RespondWithMessage(session, interaction, "FredBoard is not connected to any voice channels")
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	err := voice.StopSourceAndRemoveVoiceConn(interaction.GuildID)
	if err != nil {
		logger.Error("Leave() received an error while calling voice.StopSourceAndRemoveVoiceConn()", "error", err)
	}

	err = voiceConn.Disconnect()
	if err != nil {
		logger.Error("failed to disconnect", "error", err)

		err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	err = interactions.RespondWithMessage(session, interaction, "FredBoard left.")
	if err != nil {
		logger.Error("failed to respond to interaction", "error", err)
	}
}
