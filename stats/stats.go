package stats

import "sync"

type Counter struct {
	m     sync.RWMutex
	count int
}

func (c *Counter) Inc(x int) {
	c.m.Lock()
	defer c.m.Unlock()

	c.count += x
}

func (c *Counter) Get() int {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.count
}
