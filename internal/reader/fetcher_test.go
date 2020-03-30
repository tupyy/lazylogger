package reader

import (
	"errors"
	"testing"
	"time"
)

// Mock file reader
type MockFileReader struct {
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

func (m *MockFileReader) FetchSize() (int32, error, error) {
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

func (m *MockFileReader) GetSize() int32 {
	return int32(m.size)
}

func (m *MockFileReader) SetSize(s int32) {
	m.size = int(s)
}

func (m *MockFileReader) Close() {
	// Nothing to do
}

func (m *MockFileReader) Rewind() {
	// TODO rewind the filereader
}

func (m *MockFileReader) HasNextChunk() bool {
	return m.byteRead < m.size
}

func (m *MockFileReader) ReadNextChunk() ([]byte, error, error) {
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

type MockDataWriter struct {
	data   []byte
	err    error
	stderr error
}

func (m *MockDataWriter) WriteData(data []byte) {
	m.data = append(m.data, data...)
}

func (m *MockDataWriter) Error(stderr, err error) {
	m.stderr = stderr
	m.err = err
}

func TestFetcherNominal(t *testing.T) {
	mock := MockFileReader{
		isSizeInvalid:   false,
		isClientInvalid: false,
		fileSizeCount:   2,
		size:            0,
		byteRead:        0,
		maxChunkSize:    2,
	}

	mockDataWrite := MockDataWriter{
		data: []byte{}}

	fetcher := newFetcher(0)
	go fetcher.fetch(&mock, &mockDataWrite)

	<-time.After(3 * time.Second)
	fetcher.close()

	// As fileSizeCount=2 we should get 4 bytes of data...
	if len(mockDataWrite.data) != 4 {
		t.Errorf("Expected length: 4. Actual length: %d", len(mockDataWrite.data))
	}
}

func TestFetcherError(t *testing.T) {
	mock := MockFileReader{
		isSizeInvalid:   true,
		isClientInvalid: false,
		fileSizeCount:   2,
		size:            0,
		byteRead:        0,
		maxChunkSize:    2,
	}

	mockDataWrite := MockDataWriter{
		data: []byte{}}

	fetcher := newFetcher(0)
	go fetcher.fetch(&mock, &mockDataWrite)

	<-time.After(2 * time.Second)
	fetcher.close()

	if mockDataWrite.stderr == nil {
		t.Error("Expected stderr != ni. Actual is nil")
	}
}

func TestFetcherError2(t *testing.T) {
	mock := MockFileReader{
		isSizeInvalid:   false,
		isClientInvalid: true,
		fileSizeCount:   2,
		size:            0,
		byteRead:        0,
		maxChunkSize:    2,
	}

	mockDataWrite := MockDataWriter{
		data: []byte{}}

	fetcher := newFetcher(0)
	go fetcher.fetch(&mock, &mockDataWrite)

	<-time.After(2 * time.Second)
	fetcher.close()

	if mockDataWrite.err == nil {
		t.Error("Expected err != nil. Actual is nil")
	}
}
