package commands

import (
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Leave(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	/*
		logger := log.NewChildLogger()
		defer logger.Close()

		logger.SetData("session", &session)
		logger.SetData("interaction", &interaction)

		voiceConn, ok := session.VoiceConnections[interaction.GuildID]
		if !ok {
			logger.Info("command /leave executed without any active voice connections in guild")

			err := interactions.RespondWithMessage(session, interaction, "FredBoard is not connected to any voice channels")
			if err != nil {
				logger.ErrorWithErr("failed to respond to interaction", err)
			}

			return
		}

		source, ok := voice.VoiceConnSources[interaction.GuildID]
		if ok {
			err := source.Stop()

			if err != nil {
				logger.ErrorWithErr("error while stopping audio source", err)
			}

			delete(voice.VoiceConnSources, voiceConn.GuildID)
		}

		err := voiceConn.Disconnect()
		if err != nil {
			logger.ErrorWithErr("failed to disconnect", err)

			err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
			if err != nil {
				logger.ErrorWithErr("failed to respond to interaction", err)
			}

			return
		}

		err = interactions.RespondWithMessage(session, interaction, "FredBoard left.")
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}
	*/
}
