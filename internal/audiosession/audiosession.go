package audiosession

import (
	"fmt"
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/audio/codecs"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/ioext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

var allAudioSessions []*AudioSession = make([]*AudioSession, 0)

type audioInputState byte

const (
	audioInputState_Running = iota
	audioInputState_Paused
)

type audioInput interface {
	Subgraph() audio.Node
	State() audioInputState
	Pause()
	Resume()
	Stop()
}

type ytdlpAudioInput struct {
	subgraph audio.Node
	state    audioInputState
}

func (i *ytdlpAudioInput) Subgraph() audio.Node {
	return i.subgraph
}

func (i *ytdlpAudioInput) State() audioInputState {
	return i.state
}

func (i *ytdlpAudioInput) Pause() {
	// TODO
	panic("unimplemeted")
}

func (i *ytdlpAudioInput) Resume() {
	// TODO
	panic("unimplemeted")
}

func (i *ytdlpAudioInput) Stop() {
	// TODO
	panic("unimplemeted")
}

type audioOutput interface {
	Subgraph() audio.Node
}

type discordVoiceConn struct {
	subgraph audio.Node
	conn     *discordgo.VoiceConnection
}

func (conn *discordVoiceConn) Subgraph() audio.Node {
	return conn.subgraph
}

type AudioSession struct {
	sync.Mutex

	logger     *logging.Logger
	inputs     []audioInput
	outputs    []audioOutput
	rootMixer  *audio.MixerNode
	audioGraph *audio.Graph
	closeChan  chan struct{}
}

func (s *AudioSession) AddDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) error {
	s.Lock()
	defer s.Unlock()

	opusSendWriter := ioext.NewChannelWriter(conn.OpusSend)
	opusEncoderWriter, err := codecs.NewOpusEncoderWriter(opusSendWriter, config.Get().Audio.NumChannels, config.Get().Audio.SampleRateHz, 960) // TODO: move 960 to config file
	if err != nil {
		return fmt.Errorf("failed to create opus encoder writer: %w", err)
	}

	opusSendNode := audio.NewWriterNode(s.logger, opusEncoderWriter)
	s.audioGraph.AddNode(opusSendNode)
	s.audioGraph.CreateConnection(s.rootMixer, opusSendNode)

	s.outputs = append(s.outputs, &discordVoiceConn{subgraph: opusSendNode, conn: conn})

	return nil
}

func (s *AudioSession) RemoveDiscordVoiceConnOutput(conn *discordgo.VoiceConnection) {
	s.Lock()
	defer s.Unlock()

	idx := slices.IndexFunc(s.outputs, func(sg audioOutput) bool { vc, ok := sg.(*discordVoiceConn); return ok && vc.conn == conn })
	subgraph := s.outputs[idx]
	s.outputs = slices.Delete(s.outputs, idx, idx+1)

	s.audioGraph.RemoveNode(subgraph.Subgraph())
}

func (s *AudioSession) AddYtdlpInput(url string, quality ytdlp.YtdlpAudioQuality) (audioInput, error) {
	s.Lock()
	defer s.Unlock()

	videoReader, err, _ := ytdlp.NewVideoReader(s.logger, ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile}, url, quality)

	if err != nil {
		return nil, fmt.Errorf("failed to create video reader: %w", err)
	}

	transcoder, err, _ := ffmpeg.NewTranscoder(
		s.logger,
		ffmpeg.Config{ExePath: config.Get().Ffmpeg.ExePath},
		videoReader,
		ffmpeg.Format_PCMSigned16BitLittleEndian,
		config.Get().Audio.SampleRateHz,
		config.Get().Audio.NumChannels,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transcoder: %w", err)
	}

	// TODO: Put 0x8000 in config
	videoReaderNode := audio.NewReaderNode(s.logger, transcoder, 0x8000)

	s.audioGraph.AddNode(videoReaderNode)
	s.audioGraph.CreateConnection(videoReaderNode, s.rootMixer)

	input := &ytdlpAudioInput{subgraph: videoReaderNode, state: audioInputState_Running}
	s.inputs = append(s.inputs, input)

	return input, nil
}

func (s *AudioSession) RemoveInput(input audioInput) {
	s.Lock()
	defer s.Unlock()

	s.inputs = slices.DeleteFunc(s.inputs, func(i audioInput) bool { return i == input })
	s.audioGraph.RemoveNode(input.Subgraph())
}

func (s *AudioSession) StartTicking() {
	go func() {
		for {
			select {
			case <-s.closeChan:
				fmt.Println("==========DONE TICKING=============")
				return
			default:
				s.Lock()
				// TODO: Check if there are any outputs. If there are not, destroy the graph
				s.audioGraph.Tick()
				s.Unlock()
			}
		}
	}()
}

func (s *AudioSession) StopTicking() {
	close(s.closeChan)
}

func New(logger *logging.Logger) *AudioSession {
	rootMixer := audio.NewMixerNode(logger)

	audioGraph := audio.NewGraph(logger)
	audioGraph.AddNode(rootMixer)

	audioSession := AudioSession{
		logger:     logger,
		inputs:     make([]audioInput, 0),
		outputs:    make([]audioOutput, 0),
		rootMixer:  rootMixer,
		audioGraph: audio.NewGraph(logger),
		closeChan:  make(chan struct{}),
	}

	allAudioSessions = append(allAudioSessions, &audioSession)

	return &audioSession
}
