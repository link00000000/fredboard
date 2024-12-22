package commands

import (
	"errors"
	"fmt"
	"strings"

	"accidentallycoded.com/fredboard/v3/codecs"
	"accidentallycoded.com/fredboard/v3/discord/interactions"
	"accidentallycoded.com/fredboard/v3/discord/voice"
	"accidentallycoded.com/fredboard/v3/sources"
	"accidentallycoded.com/fredboard/v3/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

var ErrUnknownEncoding = errors.New("unknown encoding")

type opusEncodingType byte

const (
	opusEncodingType_PCMS16LE opusEncodingType = iota
	opusEncodingType_DCA0
)

type fsCommandOptions struct {
	encoding opusEncodingType
	path     string
}

func getFsOpts(interaction *discordgo.Interaction) (*fsCommandOptions, error) {
	encodingStr, err := interactions.GetRequiredStringOpt(interaction, "encoding")
	if err != nil {
		return nil, fmt.Errorf("failed to get required option \"encoding\"", err)
	}

	var encoding opusEncodingType
	switch strings.ToUpper(encodingStr) {
	case "PCMS16LE":
		encoding = opusEncodingType_PCMS16LE
	case "DCA0":
		encoding = opusEncodingType_DCA0
	default:
		return nil, ErrUnknownEncoding
	}

	path, err := interactions.GetRequiredStringOpt(interaction, "path")
	if err != nil {
		return nil, fmt.Errorf("failed to get required option \"path\"", err)
	}

	return &fsCommandOptions{encoding, path}, nil
}

func FS(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

	// get command options
	opts, err := getFsOpts(interaction)
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

	// create fs source
	source := sources.NewFSSource(opts.path)

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

	// start source
	err = source.Start()
	if err != nil {
		logger.ErrorWithErr("failed to start source", err)

		err := interactions.RespondWithError(session, interaction, "Unexpected error", err)
		if err != nil {
			logger.ErrorWithErr("failed to respond to interaction", err)
		}

		return
	}

	logger.Debug("started source")

	// transcode source to sink
	switch opts.encoding {
	case opusEncodingType_PCMS16LE:
		go encoder.EncodePCMS16LE(source, sink, 960)
	case opusEncodingType_DCA0:
		go encoder.EncodeDCA0(source, sink)
	}

	// notify user that everything is OK
	err = interactions.RespondWithMessage(session, interaction, "Playing...")
	if err != nil {
		logger.ErrorWithErr("failed to respond to interaction", err)
	}

	logger.Debug("notified user that everything is OK")

	err = source.Wait()
	if err != nil {
		logger.ErrorWithErr("error while waiting for source", err)
		return
	}
}
