package reader

import (
	"encoding/binary"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type dockerMock struct {
	data              []byte
	offset            int32
	step              int
	hasContainerError bool
	hasError          bool
}

func (d *dockerMock) ContainerLogs(containerId string) ([]byte, int32, error, error) {
	d.step++

	if d.step > 1 {
		if d.hasContainerError {
			return []byte{}, 0, errors.New("container error"), nil
		} else if d.hasError {
			return []byte{}, 0, nil, errors.New("error")
		}
	}
	buff := make([]byte, binary.MaxVarintLen16)

	for _, x := range []uint64{1, 2, 3} {
		binary.PutUvarint(buff, x)
	}
	d.offset += int32(len(buff))
	d.data = append(d.data, buff...)

	return d.data, d.offset, nil, nil
}

func TestBytesReader(t *testing.T) {
	d := &dockerMock{
		data:              []byte(nil),
		offset:            0,
		step:              0,
		hasContainerError: false,
		hasError:          false,
	}

	bReader := NewBytesReader("id", d)

	var n int32
	var e1, e2 error
	var data []byte
	for i := 1; i < 5; i++ {
		n, e1, e2 = bReader.FetchSize()

		assert.Equal(t, n, int32(i*3), "expect 2 bytes")
		assert.NotNil(t, e1)
		assert.NotNil(t, e2)
		if !bReader.HasNextChunk() {
			t.Errorf("Expected: has next chunk. Actual: no next chunk")
		}

		data, e1, e2 = bReader.ReadNextChunk()
		// We are expencting chunks of 3 bytes length
		assert.Equal(t, len(data), 3, "expect 3 bytes")
	}
}

func TestBytesReader2(t *testing.T) {
	d := &dockerMock{
		data:              []byte(nil),
		offset:            0,
		step:              0,
		hasContainerError: true,
		hasError:          false,
	}

	bReader := NewBytesReader("id", d)

	var n int32
	var e1, e2 error
	var data []byte
	n, e1, e2 = bReader.FetchSize()
	assert.Equal(t, n, 3, "expect 3 bytes")
	assert.NotNil(t, e1)
	assert.NotNil(t, e2)

	if !bReader.HasNextChunk() {
		t.Errorf("Expected: has next chunk. Actual: no next chunk")
	}

	data, e1, e2 = bReader.ReadNextChunk()
	assert.Equal(t, len(data), 3, "expect 3 bytes")

	_, e1, e2 = bReader.FetchSize()
	assert.Nil(t, e1)
	assert.NotNil(t, e2)
}

func TestBytesReader3(t *testing.T) {
	d := &dockerMock{
		data:              []byte(nil),
		offset:            0,
		step:              0,
		hasContainerError: false,
		hasError:          true,
	}

	bReader := NewBytesReader("id", d)

	var n int32
	var e1, e2 error
	var data []byte
	n, e1, e2 = bReader.FetchSize()
	assert.Equal(t, n, 3, "expect 3 bytes")
	assert.NotNil(t, e1)
	assert.NotNil(t, e2)

	if !bReader.HasNextChunk() {
		t.Errorf("Expected: has next chunk. Actual: no next chunk")
	}

	data, e1, e2 = bReader.ReadNextChunk()
	assert.Equal(t, len(data), 3, "expect 3 bytes")

	_, e1, e2 = bReader.FetchSize()
	assert.NotNil(t, e1)
	assert.Nil(t, e2)
}
