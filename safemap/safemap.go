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

// Update will update a value only if it exists. It reports whether it was able to find and update using the key.
func (s *SafeMap[Key, Value]) Update(k Key, v Value) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.m[k]
	if ok {
		s.m[k] = v
	}
	return ok
}
func (s *SafeMap[Key, Value]) Filter(fn func(k Key, v Value) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range s.m {
		isStaying := fn(key, value)
		if !isStaying {
			delete(s.m, key)
		}
	}
}
func (s *SafeMap[Key, Value]) Delete(key Key) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

func New[Key comparable, Value any]() SafeMap[Key, Value] {
	return SafeMap[Key, Value]{
		m: make(map[Key]Value),
	}
}
