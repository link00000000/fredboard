package audiosession

import (
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/ioext"
	"github.com/bwmarrin/discordgo"
)

type discordVoiceConnAudioSessionOutput struct {
	*BaseAudioSessionOutput
	conn *discordgo.VoiceConnection
}

func (conn *discordVoiceConnAudioSessionOutput) Subgraph() audio.Node {
	return conn.subgraph
}

func (s *AudioSession) AddDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) (AudioSessionOutput, error) {

	opusSendWriter := ioext.NewChannelWriter(conn.OpusSend)
	opusEncoderWriter, err := codecs.NewOpusEncoderWriter(opusSendWriter, config.Get().Audio.NumChannels, config.Get().Audio.SampleRateHz, 960) // TODO: move 960 to config file
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder writer: %w", err)
	}

	opusSendNode := audio.NewWriterNode(s.logger, opusEncoderWriter)
	output := &discordVoiceConnAudioSessionOutput{BaseAudioSessionOutput: s.AddOutput(opusSendNode), conn: conn}

	return output, nil
}
