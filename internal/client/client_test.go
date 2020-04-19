package client

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tupyy/lazylogger/internal/datasource"
)

type mockDataSourceReader struct {
	data []int

	currentStep int

	size int

	done chan interface{}

	isDone bool
}

func createMockReader(data []int) *mockDataSourceReader {
	return &mockDataSourceReader{
		currentStep: -1,
		done:        make(chan interface{}),
		data:        data,
		isDone:      false,
	}
}

func (m *mockDataSourceReader) Size() (int32, error) {

	if m.isDone {
		return 0, nil
	}

	m.currentStep++
	if m.currentStep >= len(m.data) {
		return 0, nil
	}

	// if we are at the last step and there is no data to read then close the reader
	if m.currentStep == len(m.data)-1 && m.data[len(m.data)-1] <= 0 {
		m.isDone = true
		return 0, nil
	}

	switch m.data[m.currentStep] {
	case -1:
		return 0, fmt.Errorf("read size: %w", datasource.ErrRead)
	case -2:
		m.isDone = true
		return 0, fmt.Errorf("client error: %w", datasource.ErrDatasource)
	default:
		m.size += m.data[m.currentStep]
	}

	return int32(m.size), nil
}

func (m *mockDataSourceReader) ReadAt(p []byte, off int64) (int, error) {
	var one = uint8(1)

	if m.isDone {
		return 0, nil
	}

	if off > int64(m.size) {
		return 0, fmt.Errorf("Offset bigger than size: %w", datasource.ErrRead)
	} else if off == int64(m.size) {
		return 0, io.EOF
	}

	n := int64(m.size) - off
	b := bytes.Repeat([]byte{one}, int(n))
	copied := copy(p, b)
	if int64(copied) < n {
		return copied, io.EOF
	}

	if m.currentStep >= len(m.data)-1 {
		m.isDone = true
	}

	return copied, nil
}

func (m *mockDataSourceReader) Close() error {
	return nil
}

func (m *mockDataSourceReader) ExpectedSize() int {
	var totalSize int
	for _, b := range m.data {
		val := int(b)
		if val == -2 {
			break
		}
		if val > 0 {
			totalSize += val
		}
	}

	return totalSize
}

func TestNominal(t *testing.T) {

	testData := [][]int{
		{0, 0, -2, 0, 0},
		{1, -2, 0, 0, 1},
		{0, 0, 0, 0, 0},
		{2, 5, 0, 1, 2},
		{1, 1, 1, 1, 1},
		{2, 3, 0, 1, 2},
		{2, 2, -1, 0, 0},
		{0, 0, -1, 0, 0},
		{2, 0, -1, 2, 2},
		{2, 0, -1, -1, -1},
		{-1, -1, -1, -1, -1},
		{2, 2, 9, 4, -1},
	}

	for idx, d := range testData {
		fmt.Printf("Test data set %d\n", idx)
		m := createMockReader(d)

		var client *Client
		done := make(chan struct{}, 1)
		go func() {
			client = NewFileClient(0, m)
			client.Start()

			for {
				<-time.Tick(500 * time.Millisecond)
				if m.isDone {
					done <- struct{}{}
					return
				}
			}
		}()
		<-done

		client.Stop()
		if m.ExpectedSize() == 0 {
			b := make([]byte, 2)
			n, err := client.Read(b)
			assert.EqualError(t, io.EOF, err.Error(), fmt.Sprintf("Data set %d", idx))
			assert.Equal(t, m.ExpectedSize(), n, fmt.Sprintf("Data set %d", idx))
		} else {
			b := make([]byte, m.ExpectedSize())
			n, err := client.Read(b)
			assert.Nil(t, err, fmt.Sprintf("Data set %d", idx))
			assert.Equal(t, m.ExpectedSize(), n, fmt.Sprintf("Data set %d", idx))
		}
	}
}
