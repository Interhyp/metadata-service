package cacheable

import (
	"sort"
	"sync"
)

type Impl struct {
	mu sync.Mutex

	timestamp       string
	sortedKeysCache *[]string
	values          map[string]*interface{}
}

func New() Cacheable {
	c := &Impl{
		values: make(map[string]*interface{}),
	}
	// during initial creation we don't need to lock
	c.buildKeysCacheMustHaveLock()
	return c
}

func (c *Impl) buildKeysCacheMustHaveLock() {
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	c.sortedKeysCache = &keys
}

func (c *Impl) GetTimestamp() string {
	return c.timestamp
}

func (c *Impl) SetTimestamp(timestamp string) {
	c.timestamp = timestamp
}

func (c *Impl) GetSortedKeys() *[]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sortedKeysCache
}

func (c *Impl) GetEntryRef(key string) *interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	ref, ok := c.values[key]
	if ok {
		return ref
	} else {
		// this is actually the same as the ok branch, because the zero value of a pointer is nil,
		// but writing it down for clarity's sake
		return nil
	}
}

func (c *Impl) UpdateEntryRef(key string, newRef *interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.values[key]
	if ok {
		if newRef != nil {
			// update existing
			c.values[key] = newRef
		} else {
			// delete
			delete(c.values, key)
			c.buildKeysCacheMustHaveLock()
		}
	} else {
		if newRef != nil {
			// insert new
			c.values[key] = newRef
			c.buildKeysCacheMustHaveLock()
		}
	}
}
