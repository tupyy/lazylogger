package client

import (
	"fmt"
	"io"
	"sync"
)

const (

	// MaxCacheSize how much data we keep from a logger. Set to 300kb
	MaxCacheSize = 3 * 1024 * 100
)

// Cache holds the data read from a datasource.
// It only holds 300kb of data rotating the data.
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

// This is not a perfect implementations of ReaderAt interface because it will not block
// if len(p) are not available.
//
// If n < len(p) is returning EOF.
// If the offset is greater than the size of available data, it returns EOF.
// if n == len(p) at the end of input source, it returns nil.
func (c *cache) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("offset invalid: %d", off)
	}

	defer c.mutex.Unlock()
	c.mutex.Lock()

	lp := int64(len(p))
	size := int64(len(c.data))

	if off >= size {
		return 0, io.EOF
	}

	if off+lp > size {
		n = int(size - off)
		copy(p, c.data[off:])
		return n, nil
	}

	copy(p, c.data[off:off+lp])

	n = int(lp)
	if n < len(p) {
		return n, io.EOF
	}

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
