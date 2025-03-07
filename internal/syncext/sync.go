package syncext

import "sync"

type SyncData[T any] struct {
	sync.Mutex
	Data T
}

func (m *SyncData[T]) Do(f func(*T)) {
	m.Lock()
	defer m.Unlock()
	f(&m.Data)
}

func (m *SyncData[T]) Set(value T) {
	m.Do(func(mm *T) { *mm = value })
}
