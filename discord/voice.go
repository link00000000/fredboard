package discord

import (
	"github.com/bwmarrin/discordgo"
)

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
