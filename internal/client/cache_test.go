package client

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEmptyCache(t *testing.T) {
	c := newCache()

	dataRead := make([]byte, 2)
	n, err := c.ReadAt(dataRead, 0)
	assert.Equal(t, 0, n, "read from empty cache")
	assert.EqualError(t, io.EOF, err.Error(), "read from empty cache")
}

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
	var zero = uint8(0)
	var one = uint8(1)

	c := newCache()

	data := make([]byte, MaxCacheSize)
	for i := range data {
		data[i] = zero
	}
	data[0] = one
	data[1] = one

	n, err := c.Write(data)
	assert.Nil(t, err)
	assert.Equal(t, n, len(data), "writing to cache")

	// write other 2 bytes at the end. The first 2 bytes should be 0 and 0.
	d := []byte{one, one}
	c.Write(d)

	dataRead := make([]byte, 2)
	c.ReadAt(dataRead, 0)
	assert.Equal(t, zero, dataRead[0], "expect 0")
	assert.Equal(t, zero, dataRead[1], "expect 0")

	dataEndRead := make([]byte, 2)
	n, err = c.ReadAt(dataEndRead, MaxCacheSize-2)
	assert.Equal(t, 2, n, "total bytes read")
	assert.Equal(t, nil, err, "nil error if we read till the end of cache")
	assert.Equal(t, one, dataEndRead[0], "read the last 2 bytes")
	assert.Equal(t, one, dataEndRead[1], "read the last 2 bytes")

	dataEOF := make([]byte, 2)
	n, err = c.ReadAt(dataEOF, MaxCacheSize)
	assert.Equal(t, 0, n, "no bytes read from the end of cache")
	assert.Equal(t, io.EOF, err, "EOF error from reading till the end of cache")

	dataEOF = make([]byte, 2)
	n, err = c.ReadAt(dataEOF, MaxCacheSize+22)
	assert.Equal(t, 0, n, "no bytes read if offset > MaxCacheSize")
	assert.Equal(t, io.EOF, err, "EOF error from reading with an offset > MaxCacheSize")
}
