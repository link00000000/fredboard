package audiosession

import (
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var allAudioSessions []*AudioSession = make([]*AudioSession, 0)

type audioInputState byte

const (
	audioInputState_Running = iota
	audioInputState_Paused
	audioInputState_Stopped
)

type AudioSessionInput interface {
	// returns the audio graph that is associated with this input
	Subgraph() audio.Node

	// returns the current state of the input and its playback
	State() audioInputState

	// pauses playback
	Pause()

	// resumes paused playback
	Resume()

	// stops playback (cannot be resumed)
	Stop()

	// returns an event emitter that will broadcast when the input is stopped
	OnStoppedEvent() *events.EventEmitter[struct{}]

	Equals(rhs AudioSessionInput) bool
	asBase() *BaseAudioSessionInput
}

type BaseAudioSessionInput struct {
	subgraph       audio.Node
	state          audioInputState
	onStoppedEvent *events.EventEmitter[struct{}]
}

func (i BaseAudioSessionInput) Subgraph() audio.Node {
	return i.subgraph
}

func (i BaseAudioSessionInput) State() audioInputState {
	return i.state
}

func (i *BaseAudioSessionInput) Pause() {
	i.state = audioInputState_Paused
}

func (i *BaseAudioSessionInput) Resume() {
	i.state = audioInputState_Running
}

func (i *BaseAudioSessionInput) Stop() {
	i.state = audioInputState_Stopped
	i.onStoppedEvent.Broadcast(struct{}{})
}

func (i *BaseAudioSessionInput) OnStoppedEvent() *events.EventEmitter[struct{}] {
	return i.onStoppedEvent
}

func (i *BaseAudioSessionInput) asBase() *BaseAudioSessionInput {
	return i
}

func (i *BaseAudioSessionInput) Equals(rhs AudioSessionInput) bool {
	return i == rhs.asBase()
}

type AudioSessionOutput interface {
	Subgraph() audio.Node

	Equals(rhs AudioSessionOutput) bool
	asBase() *BaseAudioSessionOutput
}

type BaseAudioSessionOutput struct {
	subgraph audio.Node
}

func (o BaseAudioSessionOutput) Subgraph() audio.Node {
	return o.subgraph
}

func (o *BaseAudioSessionOutput) asBase() *BaseAudioSessionOutput {
	return o
}

func (i *BaseAudioSessionOutput) Equals(rhs AudioSessionOutput) bool {
	return i == rhs.asBase()
}

type AudioSessionEvent_OnInputRemoved struct {
	InputRemoved     AudioSessionInput
	NInputsRemaining int
}

type AudioSessionEvent_OnOutputRemoved struct {
	OutputRemoved     AudioSessionOutput
	NOutputsRemaining int
}

type AudioSession struct {
	sync.Mutex

	logger     *logging.Logger
	inputs     []AudioSessionInput
	outputs    []AudioSessionOutput
	rootMixer  *audio.MixerNode
	audioGraph *audio.Graph

	inputWg  sync.WaitGroup // how many producers are still producing data
	outputWg sync.WaitGroup // how many consumers are still consuming data

	OnInputRemoved  *events.EventEmitter[AudioSessionEvent_OnInputRemoved]
	OnOutputRemoved *events.EventEmitter[AudioSessionEvent_OnOutputRemoved]
}

func (s *AudioSession) AddInput(node audio.Node) *BaseAudioSessionInput {
	s.Lock()
	defer s.Unlock()

	s.audioGraph.AddNode(node)
	s.audioGraph.CreateConnection(node, s.rootMixer)

	s.inputWg.Add(1)

	input := &BaseAudioSessionInput{subgraph: node, state: audioInputState_Running, onStoppedEvent: events.NewEventEmitter[struct{}]()}
	s.inputs = append(s.inputs, input)

	return input
}

func (s *AudioSession) RemoveInput(input AudioSessionInput) {
	func() {
		s.Lock()
		defer s.Unlock()

		s.inputWg.Done()

		s.inputs = slices.DeleteFunc(s.inputs, func(i AudioSessionInput) bool { return i.Equals(input) })
		s.audioGraph.RemoveNode(input.Subgraph())
	}()

	s.OnInputRemoved.Broadcast(AudioSessionEvent_OnInputRemoved{InputRemoved: input, NInputsRemaining: len(s.inputs)})
}

func (s *AudioSession) AddOutput(node audio.Node) *BaseAudioSessionOutput {
	s.Lock()
	defer s.Unlock()

	s.outputWg.Add(1)

	s.audioGraph.AddNode(node)
	s.audioGraph.CreateConnection(s.rootMixer, node)

	output := &BaseAudioSessionOutput{subgraph: node}
	s.outputs = append(s.outputs, output)

	return output
}

func (s *AudioSession) RemoveOutput(output AudioSessionOutput) {
	func() {
		s.Lock()
		defer s.Unlock()

		s.outputWg.Done()

		s.outputs = slices.DeleteFunc(s.outputs, func(o AudioSessionOutput) bool { return o.Equals(output) })
		s.audioGraph.RemoveNode(output.Subgraph())
	}()

	s.OnOutputRemoved.Broadcast(AudioSessionEvent_OnOutputRemoved{OutputRemoved: output, NOutputsRemaining: len(s.outputs)})
}

func (s *AudioSession) StartTicking() {
	processTick := func() /*continue*/ bool {
		s.Lock()
		defer s.Unlock()

		if len(s.inputs) == 0 && len(s.outputs) == 0 {
			return false
		}

		s.audioGraph.Tick()
		return true
	}

	for {
		if !processTick() {
			break
		}
	}
}

func New(logger *logging.Logger) *AudioSession {
	rootMixer := audio.NewMixerNode(logger)

	audioGraph := audio.NewGraph(logger)
	audioGraph.AddNode(rootMixer)

	audioSession := AudioSession{
		logger:     logger,
		inputs:     make([]AudioSessionInput, 0),
		outputs:    make([]AudioSessionOutput, 0),
		rootMixer:  rootMixer,
		audioGraph: audio.NewGraph(logger),

		OnInputRemoved:  events.NewEventEmitter[AudioSessionEvent_OnInputRemoved](),
		OnOutputRemoved: events.NewEventEmitter[AudioSessionEvent_OnOutputRemoved](),
	}

	allAudioSessions = append(allAudioSessions, &audioSession)

	return &audioSession
}
