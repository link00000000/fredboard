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

func FindDiscordVoiceConnAudioSessionOutput(conn *discordgo.VoiceConnection) (*DiscordVoiceConnAudioSessionOutput, error) {
	allAudioSessions.Lock()
	defer allAudioSessions.Unlock()

	for _, s := range allAudioSessions.Data {
		for _, o := range s.Outputs() {
			do, ok := o.(*DiscordVoiceConnAudioSessionOutput)
			if ok && do.HasConn(conn) {
				return do, nil
			}
		}
	}

	return nil, ErrOutputNotFound
}

type DiscordVoiceConnAudioSessionOutput struct {
	*BaseAudioSessionOutput
	conn *discordgo.VoiceConnection
}

func (o *DiscordVoiceConnAudioSessionOutput) Subgraph() audio.Node {
	return o.subgraph
}

func (o *DiscordVoiceConnAudioSessionOutput) HasConn(conn *discordgo.VoiceConnection) bool {
	return o.conn == conn
}

func (s *AudioSession) AddDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) (AudioSessionOutput, error) {

	opusSendWriter := ioext.NewChannelWriter(conn.OpusSend)
	opusEncoderWriter, err := codecs.NewOpusEncoderWriter(opusSendWriter, config.Get().Audio.NumChannels, config.Get().Audio.SampleRateHz, 960) // TODO: move 960 to config file
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder writer: %w", err)
	}

	opusSendNode := audio.NewWriterNode(s.logger, opusEncoderWriter)
	output := &DiscordVoiceConnAudioSessionOutput{BaseAudioSessionOutput: NewBaseAudioSessionOutput(s, opusSendNode), conn: conn}
	s.AddOutput(output)

	return output, nil
}
