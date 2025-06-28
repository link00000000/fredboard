package commands

import (
	"github.com/link00000000/fredboard/v3/internal/audiosession"
	"github.com/link00000000/fredboard/v3/internal/discord/interactions"
	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Leave(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) {
	if interactions.Acknowledge(logger, session, interaction) != nil {
		return
	}

	conn, err := interactions.FindVoiceConn(session, interaction)
	if err == interactions.ErrVoiceConnectionNotFound {
		logger.Debug("rejecting /Leave command due to no existing voice connection", "interaction", interaction)
		interactions.RespondWithMessage(logger, session, interaction, "FredBoard is not in any voice channel on this server.")
		return
	}

	if err != nil {
		logger.Error("failed to execute /Leave command due to an error while finding the discord voice connection", "interaction", interaction, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	output, err := audiosession.FindDiscordVoiceConnOutput(conn)
	if err != nil {
		logger.Error("failed to execute /Leave command due to error while finding the associated audio session output", "interaction", interaction, "conn", conn, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	output.Session().RemoveOutput(output)

	interactions.RespondWithMessage(logger, session, interaction, "Left")
	logger.Debug("completed /Leave command", "interaction", interaction, "conn", conn, "audioSession", output.Session(), "output", output)
}
