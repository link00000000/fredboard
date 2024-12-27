package voice

import (
	"accidentallycoded.com/fredboard/v3/audio/graph"
	"github.com/bwmarrin/discordgo"
)

type VoiceConn struct {
	source graph.SourceNode
	sink   graph.SinkNode

	vc *discordgo.VoiceConnection
}

// Key is GuildID
var conns map[string]*VoiceConn = make(map[string]*VoiceConn)

func ConnectAndAdd(source graph.SourceNode, sink graph.SinkNode, vc *discordgo.VoiceConnection) *VoiceConn {
	conn := &VoiceConn{source, sink, vc}
	conns[vc.GuildID] = conn

	return conn
}

func DisconnectAndRemove(guildId string) {
	delete(conns, guildId)
}

func Pause(guildId string) {
	// TODO
}
