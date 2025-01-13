package voice

import (
	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/events"
	"github.com/bwmarrin/discordgo"
)

// Map of GuildID to [VoiceSession]s
var voiceSessions map[string]*VoiceSession = make(map[string]*VoiceSession)

type VoiceSession struct {
	voiceConn *discordgo.VoiceConnection
	graph     *graph.AudioGraph

	OnSourceStoppedDelegate events.EventEmitter[struct{ graph.AudioGraphNode }] // TODO: Use a different param type
}

func NewVoiceSession(voiceConn *discordgo.VoiceConnection) *VoiceSession {
	return &VoiceSession{
		graph:     graph.NewAudioGraph(),
		voiceConn: voiceConn,
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
