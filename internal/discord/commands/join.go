package commands

import (
	"accidentallycoded.com/fredboard/v3/internal/audiosession"
	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Join(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) {
	if interactions.Acknowledge(logger, session, interaction) != nil {
		return
	}

	conn, err := interactions.FindOrCreateVoiceConn(session, interaction)
	if err == interactions.ErrVoiceConnectionAlreadyExists {
		logger.Debug("rejecting Join command due to existing discord voice connection", "interaction", interaction, "conn", conn)
		interactions.RespondWithMessage(logger, session, interaction, "FredBoard is already in a voice channel on this server.")
		return
	}

	if err != nil {
		logger.Error("failed to execute Join command due to an error while creating a new discord voice connection", "interaction", interaction, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	// TODO: ensure that audio session does not already exist for conn (it shouldnt exist, but it should still be asserted)

	audioSession := audiosession.New(logger)
	output, err := audioSession.AddDiscordVoiceConnOutput(conn)

	audioSession.OnOutputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnOutputRemoved) {
		if param.OutputRemoved == output {
			err := conn.Disconnect()
			if err != nil {
				logger.Error("an error ocurred while disconnecting discord voice connection", "interaction", interaction, "error", err)
			}
		}
	})

	if err != nil {
		logger.Error("failed to execute Join command due error while adding discord voice conn to the audio session", "interaction", interaction, "conn", conn, "audioSession", audioSession, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	interactions.RespondWithMessage(logger, session, interaction, "Joined")
	logger.Debug("completed Join command", "interaction", interaction, "conn", conn, "audioSession", audioSession, "output", output)
}
