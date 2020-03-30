package reader

import (
	"errors"
	"testing"
	"time"
)

// Mock file reader
type mockFileReader struct {
	// how many times the file size is increased
	fileSizeCount int

	// if true, an error will be sent when size is fetched
	isSizeInvalid bool

	// if true, a client error is mocked
	isClientInvalid bool

	size int

	byteRead int

	maxChunkSize int
}

func (m *mockFileReader) FetchSize() (int32, error, error) {
	if m.isSizeInvalid {
		return 0, errors.New("size error"), nil
	}

	if m.isClientInvalid {
		return 0, errors.New("size error"), errors.New("client error")
	}

	if m.fileSizeCount == 0 {
		return int32(m.size), nil, nil
	}

	fetchedSize := m.size
	if m.byteRead == m.size {
		m.fileSizeCount--
		fetchedSize += m.maxChunkSize
	}
	return int32(fetchedSize), nil, nil

}

func (m *mockFileReader) GetSize() int32 {
	return int32(m.size)
}

func (m *mockFileReader) SetSize(s int32) {
	m.size = int(s)
}

func (m *mockFileReader) Close() {
	// Nothing to do
}

func (m *mockFileReader) Rewind() {
	// TODO rewind the filereader
}

func (m *mockFileReader) HasNextChunk() bool {
	return m.byteRead < m.size
}

func (m *mockFileReader) ReadNextChunk() ([]byte, error, error) {
	if m.isSizeInvalid {
		return []byte{}, errors.New("size error"), nil
	}

	if m.isClientInvalid {
		return []byte{}, errors.New("size error"), errors.New("client error")
	}

	var data []byte
	data = make([]byte, m.maxChunkSize)
	m.byteRead += m.maxChunkSize
	for i, _ := range data {
		data[i] = 'a'
	}
	return data, nil, nil
}

func TestLoggerNominal(t *testing.T) {
	var dataNotifications = []DataNotification{}
	out := make(chan interface{})
	done := make(chan interface{})
	var err error

	mock := mockFileReader{
		isSizeInvalid:   false,
		isClientInvalid: false,
		fileSizeCount:   2,
		size:            0,
		byteRead:        0,
		maxChunkSize:    2,
	}

	logger := NewLogger(1, out)
	go func(done chan interface{}) {
		for {
			select {
			case data := <-out:
				if n, ok := data.(DataNotification); ok {
					dataNotifications = append(dataNotifications, n)
				} else {
					err = errors.New("received something other than DataNotification struct")
				}
			case <-done:
				return
			}
		}
	}(done)

	logger.Start(&mock)
	if !logger.IsRunning() {
		t.Error("logger expected to run. it is not running")
	}

	<-time.After(3 * time.Second)
	done <- struct{}{}

	logger.Stop()
	if logger.IsRunning() {
		t.Error("logger expected to be stopped. it is running")
	}
	data, _ := logger.RequestData(0, 2)
	if len(data) != 2 {
		t.Errorf("Expected data size 2. Actual: %d", len(data))
	}

	if err != nil {
		t.Error(err)
	}

	if len(dataNotifications) != 2 {
		t.Errorf("Expected: 2 notifications. Actual: %d.", len(dataNotifications))
	}

	if dataNotifications[0].Size != 2 {
		t.Errorf("Expected: 2 bytes. Actual: %d", dataNotifications[0].Size)
	}

	if dataNotifications[1].Size != 4 || dataNotifications[1].PreviousSize != 2 {
		t.Errorf("Expected size: 4 bytes and previous size: 2. Actual size %d. Previous size: %d", dataNotifications[1].Size, dataNotifications[1].PreviousSize)
	}
}
