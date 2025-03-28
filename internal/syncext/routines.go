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
	Name() string
	Status() string

	// runs the routine. return ErrRequestTermination to request termination of all routines
	Run() error
	Terminate(force bool)
}

type BasicRoutine struct {
	name   string
	status string
	f      func(term <-chan bool) error
	term   chan bool
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

func (r *BasicRoutine) Terminate(force bool) {
	r.term <- force
}

func NewBasicRoutine(name string, f func(term <-chan bool) error) *BasicRoutine {
	return &BasicRoutine{name: name, status: "TODO", f: f, term: make(chan bool, 1)}
}

type routineManagerRoutine struct {
	Routine Routine
	Id      RoutineId
}

// runs multiple routines and blocks until all routines are complete.
type RoutineManager struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	routines []routineManagerRoutine
}

func (m *RoutineManager) generateRoutineId() RoutineId {
	if len(m.routines) == math.MaxInt32 {
		panic("out of routine ids")
	}

	for {
		id := nextId

		if nextId == math.MaxInt32 {
			nextId = 0
		} else {
			nextId++
		}

		if !slices.ContainsFunc(m.routines, func(r routineManagerRoutine) bool { return r.Id == id }) {
			return RoutineId(id)
		}
	}
}

func (m *RoutineManager) StartRoutine(routine Routine) {
	id := m.generateRoutineId()

	m.mu.Lock()
	m.routines = append(m.routines, routineManagerRoutine{routine, id})
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer func() {
			m.mu.Lock()
			m.routines = slices.DeleteFunc(m.routines, func(r routineManagerRoutine) bool { return r.Id == id })
			m.mu.Unlock()
		}()

		err := routine.Run()
		if errors.Is(err, ErrRequestTermAllRoutines) {
			go m.TerminateAllRoutines(false)
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

func (m *RoutineManager) TerminateAllRoutines(force bool) {
	for _, v := range m.routines {
		v.Routine.Terminate(force)
	}
}

func NewRoutineManager() *RoutineManager {
	return &RoutineManager{routines: make([]routineManagerRoutine, 0)}
}
