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

	audioSession, output, exists, err := interactions.FindOrCreateAudioSession(logger, session, interaction)
	if err != nil {
		logger.Error("failed to execute /Yt command due to failure while finding or creating audio session", "interaction", interaction, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return
	}

	input, err := audioSession.AddYtdlpInput(opts.url, ytdlp.YtdlpAudioQuality_BestAudio)
	if err != nil {
		logger.Error("failed to execute /Yt command due error while adding ytdlp input to the audio session", "interaction", interaction, "audioSession", audioSession, "error", err)
		interactions.RespondWithError(logger, session, interaction, err)
		return

		// TODO: destroy audio session if it was just created
	}

	input.OnStoppedEvent().AddDelegate(func(struct{}) {
		audioSession.RemoveInput(input)
	})

	if !exists {
		audioSession.OnInputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnInputRemoved) {
			if param.NInputsRemaining == 0 {
				logger.Debug("removing discord voice conn due to all inputs to the audio session being removed", "audioSession", audioSession, "output", output)
				audioSession.RemoveOutput(output)
			}
		})
	}

	audioSession.OnOutputRemoved.AddDelegate(func(param audiosession.SessionEvent_OnOutputRemoved) {
		if param.OutputRemoved == output {
			err := output.Conn.Disconnect()
			if err != nil {
				logger.Error("an error ocurred while disconnecting discord voice connection", "interaction", interaction, "error", err)
			}
		}
	})

	if audioSession.State() == audiosession.SessionState_NotTicking {
		go audioSession.StartTicking()
	}

	interactions.RespondWithMessage(logger, session, interaction, "Playing...")
	logger.Debug("completed /Yt command", "interaction", interaction, "audioSession", audioSession, "input", input, "output", output)
}
