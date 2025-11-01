package safemap

import "sync"

type SafeMap[Key comparable, Value any] struct {
	mu sync.RWMutex
	m  map[Key]Value
}

func (s *SafeMap[Key, Value]) Get(k Key) (Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[k]
	return v, ok
}
func (s *SafeMap[Key, Value]) Set(k Key, v Value) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = v
}

func New[Key comparable, Value any]() SafeMap[Key, Value] {
	return SafeMap[Key, Value]{
		m: make(map[Key]Value),
	}
}
