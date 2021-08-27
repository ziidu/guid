package guid

import "sync"

// Cache is a security for multi gorountings
type cache struct {
	mu    sync.RWMutex
	cache map[string]*segmentBuffer
}

func (c *cache) get(bizTag string) (*segmentBuffer, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	buffer, ok := c.cache[bizTag]
	return buffer, ok
}

func (c *cache) put(bizTag string, buffer *segmentBuffer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string]*segmentBuffer)
	}
	c.cache[bizTag] = buffer
}
