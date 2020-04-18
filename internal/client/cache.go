package client

import (
	"fmt"
	"sync"
)

const (

	// MaxCacheSize how much data we keep from a logger. Set to 300kb
	MaxCacheSize = 3 * 1024 * 100
)

// Cache holds the data from a fetcher.
type cache struct {
	mutex *sync.Mutex
	data  []byte
	size  int64
}

func newCache() *cache {
	c := &cache{
		mutex: &sync.Mutex{},
		data:  []byte{},
		size:  0,
	}

	return c
}

func (c *cache) clear() {
	c.size = 0
	c.data = []byte{}
}

// Implement the ReadAt interface
func (c *cache) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("offset invalid: %d", off)
	}

	defer c.mutex.Unlock()
	c.mutex.Lock()

	lp := int64(len(p))

	if off >= c.size {
		return 0, nil
	}

	if off+lp > c.size {
		n = int(c.size - off)
		copy(p, c.data[off:])
		return n, nil
	}

	copy(p, c.data[off:off+lp])

	n = int(lp)
	return n, nil
}

// Write always writes len(p) because is removing data from
// the beginning of slice to make place for the new data
func (c *cache) Write(p []byte) (n int, err error) {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if len(c.data)+len(p) > MaxCacheSize {
		c.data = append(c.data, p...)
		bytesToRemove := len(c.data) - MaxCacheSize
		c.data = c.data[bytesToRemove:]
	} else {
		c.data = append(c.data, p...)
	}

	c.size = int64(len(c.data))
	return len(p), nil
}
