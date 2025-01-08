package events

import "math"

type DelegateHandle int

const (
	InvalidDelegateHandle DelegateHandle = 0
)

type EventEmitter[TDelegate any] struct {
	delegates    map[DelegateHandle]TDelegate
	nextHandleId int
}

func (emitter *EventEmitter[TDelegate]) Add(cb TDelegate) DelegateHandle {
	handle := emitter.nextHandle()
	emitter.delegates[handle] = cb

	return handle
}

func (emitter *EventEmitter[TDelegate]) Remove(handle DelegateHandle) {
	delete(emitter.delegates, handle)
}

func (emitter *EventEmitter[TDelegate]) Broadcast() {
  for _, d := emitter.delegates {
    d()
  }
}

func (emitter *EventEmitter[TDelegate]) nextHandle() DelegateHandle {
	handle := DelegateHandle(emitter.nextHandleId)

	if len(emitter.delegates) == math.MaxInt {
		panic("no more unique ids for event emitter")
	}

	for {
		emitter.nextHandleId++

		// the integer overflowed
		if emitter.nextHandleId < int(handle) {
			emitter.nextHandleId = 1
		}

		if _, ok := emitter.delegates[DelegateHandle(emitter.nextHandleId)]; !ok {
			// emitter.nextHandleId does not correspond to an existing delegate
			break
		}
	}

	return handle
}

func NewEventEmitter[TDelegate any]() *EventEmitter[TDelegate] {
	return &EventEmitter[TDelegate]{
		delegates:    make(map[DelegateHandle]TDelegate),
		nextHandleId: 1,
	}
}
