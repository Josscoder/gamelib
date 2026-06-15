package gamelib

import "sync"

// SyncMap is a type-safe wrapper around sync.Map for common usage patterns.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func (m *SyncMap[K, V]) Store(key K, val V) { m.m.Store(key, val) }
func (m *SyncMap[K, V]) Delete(key K)       { m.m.Delete(key) }

func (m *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (m *SyncMap[K, V]) Map() map[K]V {
	result := make(map[K]V)
	m.m.Range(func(key, value any) bool {
		result[key.(K)] = value.(V)
		return true
	})
	return result
}

func (m *SyncMap[K, V]) Len() int {
	n := 0
	m.m.Range(func(_, _ any) bool { n++; return true })
	return n
}
