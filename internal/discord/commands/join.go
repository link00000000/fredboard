package commands

import (
	"github.com/link00000000/fredboard/v3/internal/audiosession"
	"github.com/link00000000/fredboard/v3/internal/discord/interactions"
	"github.com/link00000000/go-telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Join(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) {
	if interactions.Acknowledge(logger, session, interaction) != nil {
		return
	}

	conn, exists, err := interactions.FindOrCreateVoiceConn(session, interaction)
	if err != nil {
		logger.Error("failed to execute /Join command due to failure while finding or creating voice connection", "interaction", interaction, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	if exists {
		logger.Debug("rejecting /Join command due to existing discord voice connection", "interaction", interaction, "conn", conn)
		interactions.RespondWithMessage(logger, session, interaction, "FredBoard is already in a voice channel on this server.")
		return
	}

	// TODO: ensure that audio session does not already exist for conn (it shouldnt exist, but it should still be asserted)

	audioSession := audiosession.New(logger)
	output, err := audioSession.AddDiscordVoiceConnOutput(conn)

	if err != nil {
		logger.Error("failed to execute /Join command due error while adding discord voice conn output to the audio session", "interaction", interaction, "conn", conn, "audioSession", audioSession, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return

		// TODO: destroy audio session
	}

	audioSession.OnOutputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnOutputRemoved) {
		if param.OutputRemoved == output {
			err := conn.Disconnect()
			if err != nil {
				logger.Error("an error ocurred while disconnecting discord voice connection", "interaction", interaction, "error", err)
			}
		}
	})

	interactions.RespondWithMessage(logger, session, interaction, "Joined")
	logger.Debug("completed /Join command", "interaction", interaction, "conn", conn, "audioSession", audioSession, "output", output)
}
