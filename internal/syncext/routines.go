package syncext

import (
	"errors"
	"math"
	"slices"
	"sync"
)

var ErrRequestTermAllRoutines = errors.New("termination of all routines requested")

type RoutineId int32

const (
	RoutineId_Invalid RoutineId = -1
)

var nextId RoutineId = 0

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
}

type BasicRoutine struct {
	id     RoutineId
	name   string
	status string
	f      func(term <-chan bool) error
	term   chan bool
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
	return r.f(r.term)
}

func (r *BasicRoutine) Terminate(force bool, requestedBy Routine) {
	r.term <- force
}

func NewBasicRoutine(name string, f func(term <-chan bool) error) *BasicRoutine {
	return &BasicRoutine{name: name, status: "TODO", f: f, term: make(chan bool, 1)}
}

// runs multiple routines and blocks until all routines are complete.
type RoutineManager struct {
	wg       sync.WaitGroup
	routines *SyncData[[]Routine]
}

func (m *RoutineManager) StartRoutine(routine Routine) {
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
}

// caller must retain lock for m.routines
func (m *RoutineManager) generateRoutineId() RoutineId {
	if len(m.routines.Data) == math.MaxInt32 {
		panic("out of routine ids")
	}

	for {
		id := nextId

		if nextId == math.MaxInt32 {
			nextId = 0
		} else {
			nextId++
		}

		if !slices.ContainsFunc(m.routines.Data, func(r Routine) bool { return r.Id() == id }) {
			return RoutineId(id)
		}
	}
}

func (m *RoutineManager) addRoutine(r Routine) RoutineId {
	m.routines.Lock()
	defer m.routines.Unlock()

	r.SetId(m.generateRoutineId())
	m.routines.Data = append(m.routines.Data, r)
	return r.Id()
}

func (m *RoutineManager) removeRoutine(id RoutineId) {
	m.routines.Lock()
	defer m.routines.Unlock()

	m.routines.Data = slices.DeleteFunc(m.routines.Data, func(r Routine) bool { return r.Id() == id })
}

func NewRoutineManager() *RoutineManager {
	return &RoutineManager{routines: NewSyncData(make([]Routine, 0))}
}
