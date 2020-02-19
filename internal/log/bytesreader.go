package log

// Docker represents a docker client.
type Docker interface {

	// ContainerLogs returns the log of the container as []byte, the size of log.
	ContainerLogs(containerId string) ([]byte, int32, error, error)
}

// BytesReader provides an implementation of FileReader interface.
// It works with array of bytes returned by the docker client.
// Due to the fact that we don't know the size of the log before we read it, the behaviour is different then
// RemoteReader. The data is fetched in the FetchSize method and the size of the log is returned.
// Also, we have to keep the difference between the last received data and the actual data. This differece will be return when
// ReadNextChunk is called.
type BytesReader struct {

	// container id
	id string

	// Implementation of Docker interface
	client Docker

	// holds the last data read from container
	data []byte

	// offset represents the position
	offset int32
}

// NewBytesReader creates a new BytesReader.
func NewBytesReader(id string, client Docker) *BytesReader {
	return &BytesReader{id, client, []byte(nil), 0}
}

// GetSize return the number of bytes read.
func (b *BytesReader) GetSize() int32 {
	return b.offset
}

// SetSize sets the size.
func (b *BytesReader) SetSize(size int32) {
	// nothing to do. the size if the length of data
}

// HasNextChunk returns true if the size of data is greater than the offset.
func (b *BytesReader) HasNextChunk() bool {
	return b.offset < int32(len(b.data))
}

// Rewind set the offset to 0.
func (b *BytesReader) Rewind() {
	b.offset = 0
}

// ReadNextChunk return the part of data from offset to the end of bytes array.
// It return always nil errors because the data was already fetched from container.
func (b *BytesReader) ReadNextChunk() ([]byte, error, error) {
	n := int32(len(b.data))
	if n < b.offset {
		// rewind it means that the container has restarted because we received a smaller logger
		// than the last time.
		b.Rewind()
	} else if n == b.offset {
		return []byte{}, nil, nil
	}

	b.offset = n
	return b.data, nil, nil
}

// FetchSize read the log from the container and save the any data beyond offset to data field.
// Returns the size of fetched data and container error or connection error.
func (b *BytesReader) FetchSize() (int32, error, error) {
	data, n, containerErr, connErr := b.client.ContainerLogs(b.id)
	if containerErr != nil || connErr != nil {
		return 0, containerErr, connErr
	}

	if n < b.offset {
		b.Rewind()
	}
	if n > b.offset {
		//remove the previous data and keep only the difference between the last received data and the present data.
		b.data = nil
		b.data = append([]byte(nil), data[n-b.offset:]...)
	}
	return n, nil, nil
}

func Close() {
	// TO NOTHING
}
