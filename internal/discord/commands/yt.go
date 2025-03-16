package commands

import (
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/audiosession"
	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

type ytCommandOptions struct {
	url string
}

func getYtOpts(interaction *discordgo.Interaction) (*ytCommandOptions, error) {
	url, err := interactions.GetRequiredStringOpt(interaction, "url")
	if err != nil {
		return nil, fmt.Errorf("failed to get required option \"url\": %w", err)
	}

	return &ytCommandOptions{url}, nil
}

func Yt(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) {
	if interactions.Acknowledge(logger, session, interaction) != nil {
		return
	}

	opts, err := getYtOpts(interaction)
	if err != nil {
		logger.Error("failed to execute /Yt command due to failure while getting command options", "interaction", interaction, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	conn, exists, err := interactions.FindOrCreateVoiceConn(session, interaction)
	if err != nil {
		logger.Error("failed to execute /Yt command due to failure while finding or creating voice connection", "interaction", interaction, "conn", conn, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	if exists {
		logger.Debug("rejecting /Yt command due to existing discord voice connection", "interaction", interaction, "conn", conn)
		interactions.RespondWithMessage(logger, session, interaction, "FredBoard is already in a voice channel on this server.")
		return
	}

	// TODO: ensure that audio session does not already exist for conn (it shouldnt exist, but it should still be asserted)

	audioSession := audiosession.New(logger)

	input, err := audioSession.AddYtdlpInput(opts.url, ytdlp.YtdlpAudioQuality_BestAudio)
	if err != nil {
		logger.Error("failed to execute /Yt command due error while adding ytdlp input to the audio session", "interaction", interaction, "conn", conn, "audioSession", audioSession, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return

		// TODO: destroy audio session
	}

	input.OnStoppedEvent().AddDelegate(func(struct{}) {
		audioSession.RemoveInput(input)
	})

	output, err := audioSession.AddDiscordVoiceConnOutput(conn)
	if err != nil {
		logger.Error("failed to execute /Yt command due error while adding discord voice conn output to the audio session", "interaction", interaction, "conn", conn, "audioSession", audioSession, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return

		// TODO: destroy audio session
	}

	audioSession.OnInputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnInputRemoved) {
		if param.NInputsRemaining == 0 {
			logger.Debug("removing discord voice conn due to all inputs to the audio session being removed", "conn", conn, "audioSession", audioSession)
			audioSession.RemoveOutput(output)
		}
	})

	audioSession.OnOutputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnOutputRemoved) {
		if param.OutputRemoved == output {
			err := conn.Disconnect()
			if err != nil {
				logger.Error("an error ocurred while disconnecting discord voice connection", "interaction", interaction, "error", err)
			}
		}
	})

	go audioSession.StartTicking()

	interactions.RespondWithMessage(logger, session, interaction, "Playing...")
	logger.Debug("completed /Yt command", "interaction", interaction, "conn", conn, "audioSession", audioSession, "input", input, "output", output)
}
