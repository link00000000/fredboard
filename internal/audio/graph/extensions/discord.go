package extensions

import (
	"encoding/binary"
	"fmt"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"github.com/bwmarrin/discordgo"
)

var (
	_ graph.AudioGraphNode = (*DiscordSinkNode)(nil)
)

type DiscordSinkNode struct {
	conn *discordgo.VoiceConnection
}

// Implements [nodes.AudioGraphNode]
func (node *DiscordSinkNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := graph.AssertNodeIOBounds(ins, graph.NodeIOType_In, 1, 1); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	if err := graph.AssertNodeIOBounds(outs, graph.NodeIOType_Out, 0, 0); err != nil {
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

		// TODO: Cache buffer used for p?
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
