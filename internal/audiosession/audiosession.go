package audiosession

import (
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var allSessions = syncext.NewSyncData(make([]*Session, 0))

type inputState byte

const (
	inputState_Running = iota
	inputState_Paused
	inputState_Stopped
)

type Input interface {
	Session() *Session

	// returns the audio graph that is associated with this input
	Subgraph() audio.Node

	// returns the current state of the input and its playback
	State() inputState

	// pauses playback
	Pause()

	// resumes paused playback
	Resume()

	// stops playback (cannot be resumed)
	Stop()

	// returns an event emitter that will broadcast when the input is stopped
	OnStoppedEvent() *events.EventEmitter[struct{}]

	Equals(rhs Input) bool
	asBase() *BaseInput
}

type BaseInput struct {
	session        *Session
	subgraph       audio.Node
	state          inputState
	onStoppedEvent *events.EventEmitter[struct{}]
}

func (i BaseInput) Session() *Session {
	return i.session
}

func (i BaseInput) Subgraph() audio.Node {
	return i.subgraph
}

func (i BaseInput) State() inputState {
	return i.state
}

func (i *BaseInput) Pause() {
	i.state = inputState_Paused
}

func (i *BaseInput) Resume() {
	i.state = inputState_Running
}

func (i *BaseInput) Stop() {
	i.state = inputState_Stopped
	i.onStoppedEvent.Broadcast(struct{}{})
}

func (i *BaseInput) OnStoppedEvent() *events.EventEmitter[struct{}] {
	return i.onStoppedEvent
}

func (i *BaseInput) asBase() *BaseInput {
	return i
}

func (i *BaseInput) Equals(rhs Input) bool {
	return i == rhs.asBase()
}

func NewBaseInput(session *Session, subgraph audio.Node) *BaseInput {
	return &BaseInput{
		session:        session,
		subgraph:       subgraph,
		state:          inputState_Running,
		onStoppedEvent: events.NewEventEmitter[struct{}](),
	}
}

type Output interface {
	Session() *Session
	Subgraph() audio.Node

	Equals(rhs Output) bool
	asBase() *BaseOutput
}

type BaseOutput struct {
	session  *Session
	subgraph audio.Node
}

func (o BaseOutput) Session() *Session {
	return o.session
}

func (o BaseOutput) Subgraph() audio.Node {
	return o.subgraph
}

func (o *BaseOutput) asBase() *BaseOutput {
	return o
}

func (i *BaseOutput) Equals(rhs Output) bool {
	return i == rhs.asBase()
}

func NewBaseOutput(session *Session, subgraph audio.Node) *BaseOutput {
	return &BaseOutput{
		session:  session,
		subgraph: subgraph,
	}
}

type SessionEvent_OnInputRemoved struct {
	InputRemoved     Input
	NInputsRemaining int
}

type SessionEvent_OnOutputRemoved struct {
	OutputRemoved     Output
	NOutputsRemaining int
}

type SessionState byte

const (
	SessionState_NotTicking = iota
	SessionState_Ticking
)

type Session struct {
	sync.Mutex

	logger     *logging.Logger
	inputs     []Input
	outputs    []Output
	rootMixer  *audio.MixerNode
	audioGraph *audio.Graph
	state      SessionState

	OnInputRemoved  *events.EventEmitter[SessionEvent_OnInputRemoved]
	OnOutputRemoved *events.EventEmitter[SessionEvent_OnOutputRemoved]
}

func (s *Session) AddInput(input Input) {
	s.Lock()
	defer s.Unlock()

	s.audioGraph.AddNode(input.Subgraph())
	s.audioGraph.CreateConnection(input.Subgraph(), s.rootMixer)
	s.inputs = append(s.inputs, input)
}

func (s *Session) RemoveInput(input Input) {
	func() {
		s.Lock()
		defer s.Unlock()

		s.inputs = slices.DeleteFunc(s.inputs, func(i Input) bool { return i.Equals(input) })
		s.audioGraph.RemoveNode(input.Subgraph())
	}()

	s.OnInputRemoved.Broadcast(SessionEvent_OnInputRemoved{InputRemoved: input, NInputsRemaining: len(s.inputs)})
}

func (s *Session) Inputs() []Input {
	s.Lock()
	defer s.Unlock()

	return s.inputs[:]
}

func (s *Session) AddOutput(output Output) {
	s.Lock()
	defer s.Unlock()

	s.audioGraph.AddNode(output.Subgraph())
	s.audioGraph.CreateConnection(s.rootMixer, output.Subgraph())
	s.outputs = append(s.outputs, output)
}

func (s *Session) RemoveOutput(output Output) {
	func() {
		s.Lock()
		defer s.Unlock()

		s.outputs = slices.DeleteFunc(s.outputs, func(o Output) bool { return o.Equals(output) })
		s.audioGraph.RemoveNode(output.Subgraph())
	}()

	s.OnOutputRemoved.Broadcast(SessionEvent_OnOutputRemoved{OutputRemoved: output, NOutputsRemaining: len(s.outputs)})
}

func (s *Session) Outputs() []Output {
	s.Lock()
	defer s.Unlock()

	return s.outputs[:]
}

func (s *Session) State() SessionState {
	return s.state
}

func (s *Session) StartTicking() {
	processTick := func() /*continue*/ bool {
		s.Lock()
		defer s.Unlock()

		if len(s.inputs) == 0 {
			return false
		}

		s.audioGraph.Tick()
		return true
	}

	s.Lock()
	if s.state == SessionState_Ticking {
		s.Unlock()
		return
	}

	s.state = SessionState_Ticking
	s.Unlock()

	for {
		if !processTick() {
			s.Lock()
			s.state = SessionState_NotTicking
			s.Unlock()

			break
		}
	}
}

func New(logger *logging.Logger) *Session {
	rootMixer := audio.NewMixerNode(logger)

	audioGraph := audio.NewGraph(logger)
	audioGraph.AddNode(rootMixer)

	audioSession := Session{
		logger:     logger,
		inputs:     make([]Input, 0),
		outputs:    make([]Output, 0),
		rootMixer:  rootMixer,
		audioGraph: audio.NewGraph(logger),
		state:      SessionState_NotTicking,

		OnInputRemoved:  events.NewEventEmitter[SessionEvent_OnInputRemoved](),
		OnOutputRemoved: events.NewEventEmitter[SessionEvent_OnOutputRemoved](),
	}

	func() {
		allSessions.Lock()
		defer allSessions.Unlock()

		allSessions.Data = append(allSessions.Data, &audioSession)
	}()

	return &audioSession
}
