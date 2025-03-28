package syncext

import (
	"errors"
	"math"
	"slices"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/events"
)

var ErrRequestTermAllRoutines = errors.New("termination of all routines requested")
var ErrRequestForceTermAllRoutines = errors.New("force termination of all routines requested")

type RoutineId uint32

const (
	RoutineId_Invalid RoutineId = 0
)

var nextId RoutineId = 1

type Routine interface {
	Id() RoutineId
	SetId(id RoutineId)

	Name() string
	Status() string

	// runs the routine. return ErrRequestTermination to request termination of all routines
	// blocks until the routine is complete
	Run() error

	// notify the routine that it should terminate.
	// this should never block
	Terminate(force bool, requestedBy Routine)

	// blocks until the routine is complete
	Wait()
}

type BasicRoutine struct {
	id        RoutineId
	name      string
	status    string
	f         func(term <-chan bool) error
	term      chan bool
	doneEvent *events.EventEmitter[struct{}]
}

func (r BasicRoutine) Id() RoutineId {
	return r.id
}

func (r *BasicRoutine) SetId(id RoutineId) {
	r.id = id
}

func (r BasicRoutine) Name() string {
	return r.name
}

func (r BasicRoutine) Status() string {
	return r.status
}

func (r *BasicRoutine) SetStatus(status string) {
	r.status = status
}

func (r *BasicRoutine) Run() error {
	defer r.doneEvent.Broadcast(struct{}{})
	return r.f(r.term)
}

func (r *BasicRoutine) Wait() {
	done := make(chan struct{})
	handle := r.doneEvent.AddChan(done)
	<-done
	r.doneEvent.RemoveDelegate(handle)
}

func (r *BasicRoutine) Terminate(force bool, requestedBy Routine) {
	r.term <- force
}

func NewBasicRoutine(name string, f func(term <-chan bool) error) *BasicRoutine {
	return &BasicRoutine{
		id:        RoutineId_Invalid,
		name:      name,
		status:    "TODO",
		f:         f,
		term:      make(chan bool, 1),
		doneEvent: events.NewEventEmitter[struct{}](),
	}
}

// runs multiple routines and blocks until all routines are complete.
type RoutineManager struct {
	wg       sync.WaitGroup
	routines *SyncData[[]Routine]
}

func (m *RoutineManager) StartRoutine(routine Routine) RoutineId {
	id := m.addRoutine(routine)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer m.removeRoutine(id)

		err := routine.Run()

		if errors.Is(err, ErrRequestTermAllRoutines) {
			go m.TerminateAllRoutines(false, routine)
			return
		}

		if err != nil {
			panic(err)
			// TODO
		}
	}()

	return id
}

func (m *RoutineManager) TerminateRoutine(id RoutineId, force bool) {
	if routine := m.findRoutine(id); routine != nil {
		routine.Terminate(force, nil)
	}
}

func (m *RoutineManager) WaitForRoutine(id RoutineId) {
	if routine := m.findRoutine(id); routine != nil {
		routine.Wait()
	}
}

func (m *RoutineManager) WaitForAllRoutines() {
	m.wg.Wait()
}

func (m *RoutineManager) TerminateAllRoutines(force bool, requestedBy Routine) {
	m.routines.Lock()
	defer m.routines.Unlock()

	for _, v := range m.routines.Data {
		v.Terminate(force, requestedBy)
	}

	m.routines.Data = m.routines.Data[:0]
}

func (m *RoutineManager) generateRoutineId() RoutineId {
	m.routines.Lock()
	defer m.routines.Unlock()

	if len(m.routines.Data) == math.MaxInt32-1 {
		panic("out of routine ids")
	}

	for {
		id := nextId

		if nextId == math.MaxUint32 {
			nextId = 1
		} else {
			nextId++
		}

		if !slices.ContainsFunc(m.routines.Data, func(r Routine) bool { return r.Id() == id }) {
			return RoutineId(id)
		}
	}
}

func (m *RoutineManager) addRoutine(r Routine) RoutineId {
	r.SetId(m.generateRoutineId())

	m.routines.Lock()
	defer m.routines.Unlock()

	m.routines.Data = append(m.routines.Data, r)
	return r.Id()
}

func (m *RoutineManager) removeRoutine(id RoutineId) {
	m.routines.Lock()
	defer m.routines.Unlock()

	m.routines.Data = slices.DeleteFunc(m.routines.Data, func(r Routine) bool { return r.Id() == id })
}

func (m *RoutineManager) findRoutine(id RoutineId) Routine {
	m.routines.Lock()
	defer m.routines.Unlock()

	idx := slices.IndexFunc(m.routines.Data, func(routine Routine) bool { return routine.Id() == id })

	if idx == -1 {
		return nil
	}

	return m.routines.Data[idx]
}

func NewRoutineManager() *RoutineManager {
	return &RoutineManager{routines: NewSyncData(make([]Routine, 0))}
}
