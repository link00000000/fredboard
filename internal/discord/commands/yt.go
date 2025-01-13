package commands

import (
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

type ytCommandOptions struct {
	url string
}

func getYtOpts(interaction *discordgo.Interaction) (*ytCommandOptions, error) {
	url, err := interactions.GetRequiredStringOpt(interaction, "url")
	if err != nil {
		return nil, fmt.Errorf("failed to get required option \"url\"", err)
	}

	return &ytCommandOptions{url}, nil
}

func YT(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	YTv1(session, interaction, log)
}

func YTv2(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()
	defer logger.Close()

	// 1. get command options
	opts, err := getYtOpts(interaction)
	if err != nil {
		logger.ErrorWithErr("failed to get opts", err)

		err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}

		return
	}

	logger.SetData("opts", &opts)
	logger.Debug("got required opts")

	// TODO: Check for existing voice connection for the session

	// 2. find voice channel of interaction creator
	voiceChanId, err := interactions.FindCreatorVoiceChannelId(session, interaction)
	if err == interactions.ErrVoiceChannelNotFound {
		logger.DebugWithErr("interaction creator not in a voice channel", err)

		err := interactions.RespondWithMessage(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}

		return
	}

	if err != nil {
		logger.ErrorWithErr("failed to find interaction creator's voice channel id", err)

		err := interactions.RespondWithMessage(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}

		return
	}

	logger.SetData("voiceChannelId", voiceChanId)
	logger.Debug("found interaction creator's voice channel id")

	// 3. create voice connection
	const (
		mute = false
		deaf = true
	)
	voiceConn, err := session.ChannelVoiceJoin(interaction.GuildID, voiceChanId, mute, deaf)

	if err != nil {
		logger.ErrorWithErr("failed to join voice channel", err)

		err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}

		return
	}

	// TODO: Handle voice connection disconnect

	// 4. find or create voice session
	//voiceSession := voice.FindOrCreateSession(voiceConn)

	// TODO: Add discord sink to the audio graph

	logger.SetData("voiceConn", &voiceConn)
	logger.Debug("joined voice channel of interaction creator")

	// 4. Add YT source to the audio graph
}

func YTv1(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	/*
		logger := log.NewChildLogger()
		defer logger.Close()

			logger.SetData("session", &session)
			logger.SetData("interaction", &interaction)

			// get command options
			opts, err := getYtOpts(interaction)
			if err != nil {
				logger.ErrorWithErr("failed to get opts", err)

				err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			logger.SetData("opts", &opts)
			logger.Debug("got required opts")

			existingVoiceConn, ok := session.VoiceConnections[interaction.GuildID]
			if ok {
				logger.SetData("existingVoiceConn", existingVoiceConn)
				logger.Info("voice connection already active for guild, rejecting command")

				err := interactions.RespondWithMessage(session, interaction, "FredBoard is already in a voice channel in this guild. Wait until FredBoard has left and try again.")
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			// create opus encoder
			encoder, err := codecs.NewOpusEncoder(48000, 2)
			if err != nil {
				logger.ErrorWithErr("failed to create opus encoder", err)

				err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			logger.SetData("encoder", &encoder)
			logger.Debug("created encoder")

			// create youtube source
			//source, err := sources.NewYouTubeSource(opts.url, sources.YOUTUBESTREAMQUALITY_BEST, logger)
			var source io.Reader
			if err != nil {
				logger.ErrorWithErr("failed to create youtube source", err)

				err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			logger.SetData("source", &source)
			logger.Debug("set source")

			// find voice channel
			vc, err := interactions.FindCreatorVoiceChannelId(session, interaction)

			if err == interactions.ErrVoiceChannelNotFound {
				logger.DebugWithErr("interaction creator not in a voice channel", err)

				err := interactions.RespondWithMessage(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			if err != nil {
				logger.ErrorWithErr("failed to find interaction creator's voice channel id", err)

				err := interactions.RespondWithMessage(session, interaction, "You must be in a voice channel to use this command. Join a voice channel and try again.")
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
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
				logger.ErrorWithErr("failed to join voice channel", err)

				err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			logger.SetData("voiceConn", &voiceConn)
			logger.Debug("joined voice channel of interaction creator")

			defer func() {
				err := voiceConn.Disconnect()
				if err != nil {
					logger.ErrorWithErr("failed to close voice connection", err)
					return
				}

				logger.Debug("closed voice connection")
			}()

			// create sink
			sink := voice.NewVoiceWriter(voiceConn)

			logger.SetData("sink", &sink)
			logger.Debug("created sink")

			// start source
			//err = source.Start()
			if err != nil {
				logger.ErrorWithErr("failed to start source", err)

				err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
				if err != nil {
					logger.ErrorWithErr("failed to respond to interaction", err)
				}

				return
			}

			//defer source.Stop()

			//voice.VoiceConnSources[voiceConn.GuildID] = source
			defer delete(voice.VoiceConnSources, voiceConn.GuildID)

			logger.Debug("started source")

			// transcode source to sink
			go encoder.EncodePCMS16LE(source, sink, 960)
			logger.Debug("started transcoding")

			// notify user that everything is OK
			err = interactions.RespondWithMessage(session, interaction, "Playing...")
			if err != nil {
				logger.ErrorWithErr("failed to respond to interaction", err)
			}

			logger.Debug("notified user that everything is OK")

			// cleanup source
			logger.Debug("waiting for source")
			//err = source.Wait()
			if err != nil {
				logger.ErrorWithErr("error while waiting for source", err)
			}

			logger.Debug("done")
	*/
}
