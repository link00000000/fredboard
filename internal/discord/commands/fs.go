package commands

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

// TODO: Move to graph module
var (
	_ graph.AudioGraphNode = (*DiscordSinkNode)(nil)
)

type DiscordSinkNode struct {
	conn *discordgo.VoiceConnection
}

func (node *DiscordSinkNode) PreTick() error {
	return nil
}

func (node *DiscordSinkNode) PostTick() error {
	return nil
}

func (node *DiscordSinkNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := graph.AssertNNodeIO(ins, "in", 1, 1); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	if err := graph.AssertNNodeIO(outs, "out", 0, 0); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	for {
		var encodedFrameSize int16
		err := binary.Read(ins[0], binary.LittleEndian, &encodedFrameSize)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return fmt.Errorf("DiscordSinkNode.Tick read error: %w", err)
		}

		// TODO: Cache buffer usedf or p?
		p, err := io.ReadAll(io.LimitReader(ins[0], int64(encodedFrameSize)))
		if err != nil {
			return fmt.Errorf("DiscordSinkNode.Tick read error: %w", err)
		}

		node.conn.OpusSend <- p
	}

	return nil
}

func NewDiscordSinkNode(conn *discordgo.VoiceConnection) *DiscordSinkNode {
	return &DiscordSinkNode{conn: conn}
}

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
	defer logger.Close()

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

	// create fs source
	sourceEOF := make(chan struct{}, 1)
	sourceNode := graph.NewFSFileSourceNode()
	sourceNode.OpenFile(opts.path)
	sourceNode.OnEOF = func() { sourceEOF <- struct{}{} }
	defer func() {
		err := sourceNode.CloseFile()
		if err != nil {
			logger.ErrorWithErr("failed to close FSFileSourceNode source", err)
		}
	}()

	logger.SetData("sourceNode", &sourceNode)
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
	transcodeNode := graph.NewOpusEncoderNode(48000, 1, time.Millisecond*20)
	sinkNode := NewDiscordSinkNode(voiceConn)

	pcmDiscordSinkNode := graph.NewCompositeNode()
	pcmDiscordSinkNode.AddNode(transcodeNode)
	pcmDiscordSinkNode.AddNode(sinkNode)
	pcmDiscordSinkNode.CreateConnection(transcodeNode, sinkNode)
	pcmDiscordSinkNode.SetInNode(transcodeNode)

	audioGraph := graph.NewAudioGraph()
	audioGraph.AddNode(sourceNode)
	audioGraph.AddNode(pcmDiscordSinkNode)
	audioGraph.CreateConnection(sourceNode, pcmDiscordSinkNode)

	// notify user that everything is OK
	err = interactions.RespondWithMessage(session, interaction, "Playing...")
	if err != nil {
		logger.ErrorWithErr("failed to respond to interaction", err)
	}
	logger.Debug("notified user that everything is OK")

loop:
	for {
		select {
		case <-sourceEOF:
			break loop
		default:
			err := audioGraph.Tick()
			if err != nil {
				logger.ErrorWithErr("error while ticking audio graph", err)
				return
			}
		}
	}

	logger.Debug("done")
}
