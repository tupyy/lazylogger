package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	c := newCache()

	data := []byte{'l', 'o', 'g', 'g', 'e', 'r'}
	n, err := c.Write(data)
	assert.Nil(t, err)
	assert.Equal(t, n, len(data), "writing to cache")

	dataRead := make([]byte, len(data))
	c.ReadAt(dataRead, 0)
	assert.Equal(t, string(dataRead), string(data), "data read error")
}

func TestReadOffset(t *testing.T) {
	c := newCache()

	data := []byte{'l', 'o', 'g', 'g', 'e', 'r'}
	n, err := c.Write(data)
	assert.Nil(t, err)
	assert.Equal(t, n, len(data), "writing to cache")

	dataRead := make([]byte, 1)
	c.ReadAt(dataRead, 2)
	assert.Equal(t, string(dataRead), "g", "expect char g")
}

func TestWrite(t *testing.T) {
	c := newCache()

	data := make([]byte, MaxCacheSize)
	for i, _ := range data {
		data[i] = 0
	}
	data[0] = 1
	data[1] = 1

	n, err := c.Write(data)
	assert.Nil(t, err)
	assert.Equal(t, n, len(data), "writing to cache")

	// write other 2 bytes at the end. The first 2 bytes should be 0 and 0.
	d := []byte{2, 2}
	c.Write(d)

	dataRead := make([]byte, 2)
	c.ReadAt(dataRead, 0)
	assert.Equal(t, dataRead[0], 0, "expect 0")
	assert.Equal(t, dataRead[1], 0, "expect 0")

	dataEndRead := make([]byte, 2)
	c.ReadAt(dataEndRead, MaxCacheSize-2)
	assert.Equal(t, dataRead[0], 2, "expect 2")
	assert.Equal(t, dataRead[1], 2, "expect 2")
}
