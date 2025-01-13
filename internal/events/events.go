package events

import "math"

type DelegateHandle int

const (
	InvalidDelegateHandle DelegateHandle = 0
)

type Delegate[TDelegateParam any] func(param TDelegateParam)

type EventEmitter[TDelegateParam any] struct {
	delegates    map[DelegateHandle](Delegate[TDelegateParam])
	nextHandleId int
}

func (emitter *EventEmitter[TDelegateParam]) Add(cb Delegate[TDelegateParam]) DelegateHandle {
	handle := emitter.nextHandle()
	emitter.delegates[handle] = cb

	return handle
}

func (emitter *EventEmitter[TDelegateParam]) Remove(handle DelegateHandle) {
	delete(emitter.delegates, handle)
}

func (emitter *EventEmitter[TDelegateParam]) Broadcast(param TDelegateParam) {
	for _, d := range emitter.delegates {
		d(param)
	}
}

func (emitter *EventEmitter[TDelegateParam]) nextHandle() DelegateHandle {
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

func NewEventEmitter[TDelegateParam any]() *EventEmitter[TDelegateParam] {
	return &EventEmitter[TDelegateParam]{
		delegates:    make(map[DelegateHandle](Delegate[TDelegateParam])),
		nextHandleId: 1,
	}
}
