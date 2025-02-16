package sessions

import (
	"fmt"
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var sessions []*Session

type SessionState byte

const (
	SessionState_Ready SessionState = iota
	SessionState_Ticking
)

type AudioGraphNodeId string

type AudioGraphNodeWithId struct {
	node graph.AudioGraphNode
	id   AudioGraphNodeId
}

func NewAudioGraphNodeId_FromDiscordGuildId(guildId string) AudioGraphNodeId {
	return AudioGraphNodeId(fmt.Sprintf("discordguild-%s", guildId))
}

type Session struct {
	OnAudioGraphNodeAdded   *events.EventEmitter[struct{ node graph.AudioGraphNode }]
	OnAudioGraphNodeRemoved *events.EventEmitter[struct{ node graph.AudioGraphNode }]
	OnDestroySession        *events.EventEmitter[struct{}]

	ShouldDestroyWhenNoSources bool

	audioGraph      *graph.AudioGraph
	sourceMixerNode graph.AudioGraphNode
	sinkTeeNode     graph.AudioGraphNode
	sourceNodes     []AudioGraphNodeWithId
	sinkNodes       []AudioGraphNodeWithId

	audioGraphMutex sync.RWMutex
	state           SessionState

	logger *logging.Logger
}

func (session *Session) GetState() SessionState {
	return session.state
}

func (session *Session) RunTickGraphLoop() {
	session.state = SessionState_Ticking

	onAudioSourceAddedChan := make(chan struct{ node graph.AudioGraphNode })
	onAudioGraphNodeAddedDelegateHandle := session.OnAudioGraphNodeAdded.AddChan(onAudioSourceAddedChan)
	defer session.OnAudioGraphNodeAdded.RemoveDelegate(onAudioGraphNodeAddedDelegateHandle)

	onDestroySessionChan := make(chan struct{})
	onDestroySessionDelegateHandle := session.OnDestroySession.AddChan(onDestroySessionChan)
	defer session.OnDestroySession.RemoveDelegate(onDestroySessionDelegateHandle)

	for {
		// Wait for either the session to be destroyed or an audio source to be added
		if len(session.sourceNodes) == 0 {
			select {
			case <-onDestroySessionChan:
				return
			case <-onAudioSourceAddedChan:
				continue
			}
		}

		session.audioGraphMutex.RLock()
		defer session.audioGraphMutex.RUnlock()

		// One final check in case the nodes did change since receiving a signal and acquiring the mutex
		if len(session.sourceNodes) == 0 {
			continue
		}

		err := session.audioGraph.Tick()

		if err != nil {
			session.logger.Error("VoiceSession.RunTickGraphLoop() error ticking audio graph: %w", "error", err)
			session.Destroy()
			return
		}
	}
}

func (session *Session) AddAudioSource(node graph.AudioGraphNode) error {
	session.audioGraphMutex.Lock()
	defer session.audioGraphMutex.Unlock()

	err := session.audioGraph.AddNode(node)
	if err != nil {
		return fmt.Errorf("VoiceSession.AddAudioSource() error adding node to graph: %w", err)
	}

	err = session.audioGraph.CreateConnection(node, session.sourceMixerNode)
	if err != nil {
		return fmt.Errorf("VoiceSession.AddAudioSource() error creating connection in graph: %w", err)
	}

	session.OnAudioGraphNodeAdded.Broadcast(struct{ node graph.AudioGraphNode }{node: node})

	return nil
}

func (session *Session) RemoveAudioSource(node graph.AudioGraphNode) error {
	session.audioGraphMutex.Lock()
	defer session.audioGraphMutex.Unlock()

	err := session.audioGraph.DestroyConnection(node, session.sourceMixerNode)
	if err != nil {
		return fmt.Errorf("VoiceSession.RemoveAudioSource() error removing connection from graph: %w", err)
	}

	err = session.audioGraph.RemoveNode(node)
	if err != nil {
		return fmt.Errorf("VoiceSession.RemoveAudioSource() error removing node from graph: %w", err)
	}

	session.OnAudioGraphNodeRemoved.Broadcast(struct{ node graph.AudioGraphNode }{node: node})

	if session.ShouldDestroyWhenNoSources && len(session.sourceNodes) == 0 {
		session.Destroy()
	}

	return nil
}

func (session *Session) Destroy() {
	session.OnDestroySession.Broadcast(struct{}{})
}

func NewVoiceSession(logger *logging.Logger) *Session {
	sourceMixerNode := graph.NewMixerNode()
	//sinkTeeNode := graph.NewTeeNode() // TODO

	audioGraph := graph.NewAudioGraph()
	audioGraph.AddNode(sourceMixerNode)
	//audioGraph.AddNode(sinkTeeNode) // TODO
	//audioGraph.CreateConnection(sourceMixerNode, sinkTeeNode) // TODO

	return &Session{
		OnAudioGraphNodeAdded:   events.NewEventEmitter[struct{ node graph.AudioGraphNode }](),
		OnAudioGraphNodeRemoved: events.NewEventEmitter[struct{ node graph.AudioGraphNode }](),
		OnDestroySession:        events.NewEventEmitter[struct{}](),

		ShouldDestroyWhenNoSources: false,

		audioGraph:      audioGraph,
		sourceMixerNode: sourceMixerNode,
		//sinkTeeNode:     sinkTeeNode, // TODO
		sourceNodes: make([]AudioGraphNodeWithId, 0),
		sinkNodes:   make([]AudioGraphNodeWithId, 0),

		state:  SessionState_Ready,
		logger: logger,
	}
}

func CreateSession(logger *logging.Logger) *Session {
	session := NewVoiceSession(logger)
	sessions = append(sessions, session)

	return session
}

func DestroySession(session *Session) {
	session.Destroy()

	idx := slices.IndexFunc(sessions, func(s *Session) bool { return s == session })
	if idx != -1 {
		slices.Delete(sessions, idx, idx+1)
	}
}

func FindSessionFunc(f func(*Session) bool) (*Session, bool) {
	idx := slices.IndexFunc(sessions, f)
	if idx == -1 {
		return nil, false
	}

	return sessions[idx], true
}
