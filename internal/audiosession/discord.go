package audiosession

import (
	"errors"
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/ioext"
	"github.com/bwmarrin/discordgo"
)

var ErrOutputNotFound = errors.New("audio session output not found")

type DiscordVoiceConnOutput struct {
	*BaseOutput
	Conn *discordgo.VoiceConnection
}

func (o *DiscordVoiceConnOutput) Subgraph() audio.Node {
	return o.subgraph
}

func (s *Session) AddDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) (*DiscordVoiceConnOutput, error) {
	opusSendWriter := ioext.NewChannelWriter(conn.OpusSend)
	opusEncoderWriter, err := codecs.NewOpusEncoderWriter(opusSendWriter, config.Get().Audio.NumChannels, config.Get().Audio.SampleRateHz, 960) // TODO: move 960 to config file
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder writer: %w", err)
	}

	opusSendNode := audio.NewWriterNode(s.logger, opusEncoderWriter)
	output := &DiscordVoiceConnOutput{BaseOutput: NewBaseOutput(s, opusSendNode), Conn: conn}
	s.AddOutput(output)

	return output, nil
}

func FindDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) (*DiscordVoiceConnOutput, error) {
	allSessions.Lock()
	defer allSessions.Unlock()

	for _, s := range allSessions.Data {
		for _, o := range s.Outputs() {
			do, ok := o.(*DiscordVoiceConnOutput)
			if ok && do.Conn == conn {
				return do, nil
			}
		}
	}

	return nil, ErrOutputNotFound
}
