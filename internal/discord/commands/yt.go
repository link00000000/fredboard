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

func YT(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()
	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

	// get command options
	opts, err := getYtOpts(interaction)
	if err != nil {
		logger.Error("failed to get opts", "error", err)

		err := interactions.RespondWithErrorMessage_NoLog(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	logger.SetData("opts", &opts)
	logger.Debug("got required opts")

	existingVoiceConn, ok := session.VoiceConnections[interaction.GuildID]
	if ok {
		logger.SetData("existingVoiceConn", existingVoiceConn)
		logger.Info("voice connection already active for guild, rejecting command")

		err := interactions.RespondWithMessage_NoLog(session, interaction, "FredBoard is already in a voice channel in this guild. Wait until FredBoard has left and try again.")
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	// find voice channel
	vc, err := interactions.FindCreatorVoiceChannelId(session, interaction)

	if err == interactions.ErrVoiceChannelNotFound {
		logger.Debug("interaction creator not in a voice channel", "error", err)

		err := interactions.RespondWithMessage_NoLog(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	if err != nil {
		logger.Error("failed to find interaction creator's voice channel id", "error", err)

		err := interactions.RespondWithMessage_NoLog(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	logger.SetData("voiceChannelId", vc)
	logger.Debug("found interaction creator's voice channel id")

	// create voice connection
	const (
		mute = false
		deaf = true
	)
	voiceConn, err := session.ChannelVoiceJoin(interaction.GuildID, vc, mute, deaf)

	if err != nil {
		logger.Error("failed to join voice channel", "error", err)

		err := interactions.RespondWithErrorMessage_NoLog(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		return
	}

	logger.SetData("voiceConn", &voiceConn)
	logger.Debug("joined voice channel of interaction creator")

	// create audio graph
	audioSession := audiosession.New(logger)
	audioSessionOutput, err := audioSession.AddDiscordVoiceConnOutput(voiceConn)
	if err != nil {
		logger.Error("failed to create discord voice conn output on audio session", "error", err)

		err := interactions.RespondWithErrorMessage_NoLog(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		// TODO: destroy audio session

		return
	}

	audioSessionInput, err := audioSession.AddYtdlpInput(opts.url, ytdlp.YtdlpAudioQuality_BestAudio)
	if err != nil {
		logger.Error("failed to create discord voice conn output on audio session", "error", err)

		err := interactions.RespondWithErrorMessage_NoLog(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.Error("failed to respond to interaction", "error", err)
		}

		// TODO: destroy audio session

		return
	}

	audioSessionInput.OnStoppedEvent().AddDelegate(func(struct{}) {
		audioSession.RemoveInput(audioSessionInput)
	})

	audioSession.OnInputRemoved.AddDelegate(func(param audiosession.AudioSessionEvent_OnInputRemoved) {
		if param.NInputsRemaining == 0 {
			logger.Debug("removing discord voice conn due to all inputs to the audio session being removed", "voiceConn", voiceConn, "audioSession", audioSession)
			audioSession.RemoveOutput(audioSessionOutput)
		}
	})

	audioSession.OnOutputRemoved.AddDelegate(func(param audiosession.AudioSessionEvent_OnOutputRemoved) {
		if param.OutputRemoved.Equals(audioSessionOutput) {
			logger.Debug("closing discord voice conn due to associated audio session output being removed", "voiceConn", voiceConn, "audioSession", audioSession, "audioSessionOutput", audioSessionOutput)

			err := voiceConn.Disconnect()
			if err != nil {
				logger.Error("an error ocurred while disconnecting discord voice conn", "voiceConn", voiceConn, "error", err)
			}
		}
	})

	go audioSession.StartTicking()

	// notify user that everything is ok
	err = interactions.RespondWithMessage_NoLog(session, interaction, "Playing...")
	if err != nil {
		logger.Error("failed to respond to interaction", "error", err)
	}

	logger.Debug("notified user that everything is OK")
	logger.Debug("done")
}
