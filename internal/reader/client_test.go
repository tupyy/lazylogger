package reader

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	steps []int32

	currentStep int32

	size int32

	done chan interface{}
}

func (m *mockReader) Size() (int32, error) {

	m.currentStep++
	if m.currentStep >= int32(len(m.steps)) {
		return 0, nil
	}

	// if we are at the last step and there is no data to read then close the reader
	if m.currentStep == int32(len(m.steps)-1) && m.steps[len(m.steps)-1] == 0 {
		m.done <- struct{}{}
		return 0, nil
	}

	switch m.steps[m.currentStep] {
	case -1:
		return 0, fmt.Errorf("read size: %w", ErrRead)
	default:
		m.size += m.steps[m.currentStep]
	}

	return m.size, nil
}

func (m *mockReader) ReadAt(p []byte, off int64) (int, error) {

	if off > int64(m.size) {
		return 0, fmt.Errorf("Offset bigger than size: %w", ErrRead)
	} else if off == int64(m.size) {
		return 0, io.EOF
	}

	n := int64(m.size) - off
	b := bytes.Repeat([]byte{1}, int(n))
	copied := copy(p, b)
	if int64(copied) < n {
		return copied, io.EOF
	}

	if m.currentStep >= int32(len(m.steps)-1) {
		m.done <- struct{}{}
	}

	return copied, nil
}

func (m *mockReader) Close() error {
	return nil
}

func TestNominal(t *testing.T) {

	m := &mockReader{
		currentStep: -1,
		steps:       []int32{1, 1, 0, 0, 1},
		done:        make(chan interface{}),
	}
	client := NewFileClient(0, m)
	client.Start()
	<-m.done
	client.Stop()

	b := make([]byte, 3)
	n, err := client.Read(b)
	assert.Nil(t, err, "read error not nil")
	assert.Equal(t, 3, n, "byte read error")

	// try to read more than 3 bytes. Expect 3 bytes.
	b = make([]byte, 10)
	n, err = client.Read(b)
	assert.Nil(t, err, "read error not nil")
	assert.Equal(t, 3, n, "byte read error")
}
