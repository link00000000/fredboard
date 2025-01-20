package voice

import (
	"sync"
	"time"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"github.com/bwmarrin/discordgo"
)

// Map of GuildID to [VoiceSession]s
var voiceSessions map[string]*VoiceSession = make(map[string]*VoiceSession)

type VoiceSession struct {
	WaitGroup sync.WaitGroup

	voiceConn       *discordgo.VoiceConnection
	audioGraph      *graph.AudioGraph
	sourceMixerNode graph.AudioGraphNode
	wgDone          chan struct{}
}

func (session *VoiceSession) TickGraphUntilComplete() {
	go func() {
		select {
		case <-session.wgDone:
			return
		default:
			session.audioGraph.Tick()
			// TODO: Handle errors
		}
	}()

	session.WaitGroup.Wait()
	session.wgDone <- struct{}{}
}

func (session *VoiceSession) AddAudioSource(node graph.AudioGraphNode) {
  session.
}

func NewVoiceSession(voiceConn *discordgo.VoiceConnection) *VoiceSession {
	sourceMixerNode := graph.NewMixerNode()
	opusEncoderNode := graph.NewOpusEncoderNode(48000, 2, time.Millisecond*20)
	discordSinkNode := graph.NewDiscordSinkNode(voiceConn)

	audioGraph := graph.NewAudioGraph()
	audioGraph.AddNode(sourceMixerNode)
	audioGraph.AddNode(opusEncoderNode)
	audioGraph.AddNode(discordSinkNode)
	audioGraph.CreateConnection(sourceMixerNode, opusEncoderNode)
	audioGraph.CreateConnection(opusEncoderNode, discordSinkNode)

	return &VoiceSession{
		audioGraph:      graph.NewAudioGraph(),
		voiceConn:       voiceConn,
		sourceMixerNode: sourceMixerNode,
	}
}

func FindOrCreateSession(voiceConn *discordgo.VoiceConnection) *VoiceSession {
	if session, ok := voiceSessions[voiceConn.GuildID]; ok {
		return session
	}

	return NewVoiceSession(voiceConn)
}

func DestroySession(guildId string) {
	if _, ok := voiceSessions[guildId]; ok {
		// TODO: Stop audio graph
		delete(voiceSessions, guildId)
	}
}
