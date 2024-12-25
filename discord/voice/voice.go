package voice

import (
	"accidentallycoded.com/fredboard/v3/sources"
	"github.com/bwmarrin/discordgo"
)

// Map of guild id to audio source
var VoiceConnSources map[string]sources.Source = make(map[string]sources.Source)

type VoiceWriter struct {
	voiceConnection *discordgo.VoiceConnection
}

func NewVoiceWriter(voiceConnection *discordgo.VoiceConnection) *VoiceWriter {
	return &VoiceWriter{voiceConnection: voiceConnection}
}

// Implements [io.Writer]
func (writer *VoiceWriter) Write(bytes []byte) (int, error) {
	writer.voiceConnection.OpusSend <- bytes
	return len(bytes), nil
}
