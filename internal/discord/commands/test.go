package commands

import (
	"time"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/discord/interactions"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

func Test(session *discordgo.Session, interaction *discordgo.Interaction, log *logging.Logger) {
	logger := log.NewChildLogger()
	defer logger.Close()

	logger.SetData("session", &session)
	logger.SetData("interaction", &interaction)

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

	// create sample1 fs source
	source1EOF := make(chan struct{}, 1)
	source1Node := graph.NewFSFileSourceNode()
	source1Node.OpenFile("./test/testdata/sample1.pcms16le")
	source1Node.OnEOF = func() { source1EOF <- struct{}{} }
	defer func() {
		err := source1Node.CloseFile()
		if err != nil {
			logger.ErrorWithErr("failed to close FSFileSourceNode source", err)
		}
	}()

	logger.SetData("sourceNode1", &source1Node)
	logger.Debug("set source 1")

	// create sample2 fs source
	source2EOF := make(chan struct{}, 1)
	source2Node := graph.NewFSFileSourceNode()
	source2Node.OpenFile("./test/testdata/sample2.pcms16le")
	source2Node.OnEOF = func() { source2EOF <- struct{}{} }
	defer func() {
		err := source2Node.CloseFile()
		if err != nil {
			logger.ErrorWithErr("failed to close FSFileSourceNode source", err)
		}
	}()

	logger.SetData("sourceNode2", &source2Node)
	logger.Debug("set source 2")

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

	// create audio graph
	transcodeNode := graph.NewOpusEncoderNode(48000, 1, time.Millisecond*20)
	sinkNode := NewDiscordSinkNode(voiceConn)

	pcmDiscordSinkNode := graph.NewCompositeNode()
	pcmDiscordSinkNode.AddNode(transcodeNode)
	pcmDiscordSinkNode.AddNode(sinkNode)
	pcmDiscordSinkNode.CreateConnection(transcodeNode, sinkNode)
	pcmDiscordSinkNode.SetInNode(transcodeNode)

	audioGraph := graph.NewAudioGraph()
	audioGraph.AddNode(source1Node)
	audioGraph.AddNode(source2Node)

	mixerNode := graph.NewMixerNode()
	audioGraph.AddNode(mixerNode)
	audioGraph.CreateConnection(source1Node, mixerNode)
	audioGraph.CreateConnection(source2Node, mixerNode)

	audioGraph.AddNode(pcmDiscordSinkNode)
	audioGraph.CreateConnection(mixerNode, pcmDiscordSinkNode)

	// notify user that everything is OK
	err = interactions.RespondWithMessage(session, interaction, "Playing...")
	if err != nil {
		logger.ErrorWithErr("failed to respond to interaction", err)
	}
	logger.Debug("notified user that everything is OK")

	nCompletedSources := 0
loop:
	for {
		select {
		case <-source1EOF:
			nCompletedSources += 1
			if nCompletedSources == 2 {
				break loop
			}
		case <-source2EOF:
			nCompletedSources += 1
			if nCompletedSources == 2 {
				break loop
			}
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
