package log

import (
	"errors"
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
	d.offset += 2

	if d.step > 1 {
		if d.hasContainerError {
			return []byte{}, 0, errors.New("container error"), nil
		} else if d.hasError {
			return []byte{}, 0, nil, errors.New("error")
		}
	}
	d.data = append(d.data, []byte("1")...)

	return d.data, d.offset, nil, nil
}

func TestNominal(t *testing.T) {
	d := &dockerMock{
		data:              []byte(nil),
		offset:            0,
		step:              0,
		hasContainerError: false,
		hasError:          false,
	}

	bReader := NewBytesReader("id", d)

	n, e1, e2 := bReader.FetchSize()
	if n != 2 {
		t.Errorf("Expected: 2. Actual: %d", n)
	}
	if e1 != nil || e2 != nil {
		t.Errorf("Expected: nil. Actual: %s %s", e1, e2)
	}

	if !bReader.HasNextChunk() {
		t.Errorf("Expected: has next chunk. Actual: no next chunk")
	}

	data1, e1, e2 := bReader.ReadNextChunk()
	if len(data1) != 2 {
		t.Errorf("Expected: len(data1) == 2. Actual: %d", len(data1))
	}

	n, e1, e2 = bReader.FetchSize()
	if n != 4 {
		t.Errorf("Expected: 2. Actual: %d", n)
	}
	if e1 != nil || e2 != nil {
		t.Errorf("Expected: nil. Actual: %s %s", e1, e2)
	}

	if !bReader.HasNextChunk() {
		t.Errorf("Expected: has next chunk. Actual: no next chunk")
	}

	data, containerErr, err := bReader.ReadNextChunk()
	if len(data) != 4 {
		t.Errorf("Expected: len(d) == 2. Actual: %d", len(data))
	}
}
