package counter

import "sync"

type Counter struct {
	mu    sync.RWMutex
	count int64
}

func (c *Counter) Increment() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	return c.count
}

func (c *Counter) Get() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}
