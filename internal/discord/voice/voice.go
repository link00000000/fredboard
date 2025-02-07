package voice

import (
	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	graph_extensions "accidentallycoded.com/fredboard/v3/internal/audio/graph/extensions"
	"github.com/bwmarrin/discordgo"
)

// Map of guild id to audio source
var VoiceConnSources map[string]graph.AudioGraphNode = make(map[string]graph.AudioGraphNode)

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

func StopSourceAndRemoveVoiceConn(guildId string) error {
	node, ok := VoiceConnSources[guildId]
	var err error = nil

	if ok {
		switch v := node.(type) {
		case *graph_extensions.YouTubeSourceNode:
			err = v.Stop()
		case *graph.FSFileSourceNode:
			err = v.CloseFile()
		}
	}

	delete(VoiceConnSources, guildId)

	return err
}
