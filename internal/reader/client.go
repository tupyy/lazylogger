package reader

import (
	"errors"
	"io"

	"github.com/golang/glog"
)

type State uint32

const (
	StateIdle State = iota

	// StateConnecting means the client is try to connect to server
	StateConnecting

	// StateDegrated means the client is still running but there is a problem with the file
	StateDegrated

	// StateError means that the client is stopped mostly due to ssh connectione erros
	StateError

	// StateStopped means the client is stopped.
	StateStopped

	// StateRunning means the client is running healthy.
	StateRunning

	defaultChunkSize = 100 * 1024 // 100k
)

// TODO write good doc
type DataSourceReader interface {
	io.ReaderAt
	io.Closer

	// Size return the size of the file and any error encountered.
	Size() (int32, error)
}

type Client interface {
	io.Reader
	io.ReaderAt

	// Start the client
	Start()

	// Stop client
	Stop()

	// Return the size of the cache
	Size() int32

	// add writer
	AddWriter(w io.Writer)

	// Remove writer
	RemoveWriter(w io.Writer)
}

// Client reads data from file and send data notification to clients.
type FileClient struct {
	Id    int
	State State

	done chan struct{}

	// cache
	cache *cache

	reader DataSourceReader

	bytesRead int32
	size      int32

	writers writers
}

// Array of io.Writer
type writers []io.Writer

// Loop through each io.Writer and write p
func (w writers) Write(p []byte) {
	for _, writer := range w {
		_, err := writer.Write(p)
		if err != nil {
			glog.V(1).Infof("Error writing to data: %s", err)
		}
	}
}

// New creates a new logger
func NewFileClient(id int, reader DataSourceReader) *FileClient {
	c := &FileClient{
		Id:      id,
		cache:   newCache(),
		done:    make(chan struct{}),
		writers: []io.Writer{},
		reader:  reader,
	}

	return c
}

// Start the client
func (c *FileClient) Start() {
	go c.fetch()
}

// Stop stop reading the file. It doesn't disconnect the client.
// it is just stop reading the file.
func (c *FileClient) Stop() {
	c.done <- struct{}{}
	c.State = StateStopped
}

// Implementation of ReaderAt interface reading from cache.
func (c *FileClient) ReadAt(p []byte, off int64) (n int, err error) {
	return c.cache.ReadAt(p, off)
}

// Implementation of Reader interface reading from cache.
func (c *FileClient) Read(p []byte) (n int, err error) {
	return c.cache.ReadAt(p, 0)
}

// Size returns the size of the cache.
func (c *FileClient) Size() int32 {
	return int32(len(c.cache.data))
}

func (c *FileClient) AddWriter(w io.Writer) {
	c.writers = append(c.writers, w)
}

// Fetch the data from reader
// It starts by fetching the size of data source. If the fetched size is greater than the old one, it will start fetching data.
// When all the data has been fetched, it starts fetching size again. So on until it stops.
func (c *FileClient) fetch() {
	fetchData := make(chan struct{})
	fetchSize := make(chan struct{})

	startFetchingSize := func() { fetchSize <- struct{}{} }
	startFetchingData := func() { fetchData <- struct{}{} }
	stop := func() { c.done <- struct{}{} }

	go startFetchingSize()
	for {
		select {
		case <-c.done:
			glog.V(2).Info("Fetch data stopped")
			return
		case <-fetchData:
			chunk := c.computeNextChunk()
			glog.V(2).Infof("Fetching next chunk of %d bytes.", chunk)

			p := make([]byte, chunk)
			bytesRead, err := c.reader.ReadAt(p, int64(c.bytesRead))
			if err != nil {
				c.State = c.handleStateChange(err)
				if c.State == StateError {
					go stop()
				} else {
					// keep fetching size even the file is not available anymore. maybe is only temporary unavailable
					go startFetchingSize()
				}
			} else {
				if bytesRead == len(p) {
					c.cache.Write(p)
					c.writers.Write(p)
				}
				c.bytesRead += int32(bytesRead)
				if c.hasNextChunk() {
					go startFetchingData()
				} else {
					go startFetchingSize()
				}
			}
		case <-fetchSize:
			newSize, err := c.reader.Size()
			if err != nil {
				c.State = c.handleStateChange(err)
				if c.State == StateError {
					go stop()
				}
			}

			// if new size is larger than old one start fetching data
			if c.size < newSize {
				c.size = newSize
				go startFetchingData()
			} else {
				go startFetchingSize()
			}
		}
	}
}

// Error of type ErrClient change the state to ERROR because they must represents error in fatal error in readers.
// Usually, this means that the data source has crashed (e.g. ssh connection ended).
// Error of type ErrRead change state to DEGRADED meaning that the reading operation failed but the connection is still ok.
func (c *FileClient) handleStateChange(err error) State {
	glog.V(2).Infof("%s", err)
	if errors.Is(err, ErrClient) {
		glog.V(1).Infof("Client: %d. Status changed to ERROR", c.Id)
		return StateError
	} else if errors.Is(err, ErrRead) {
		glog.V(1).Infof("Client %d. Status changed to DEGRADED", c.Id)
		return StateDegrated
	}
	glog.V(1).Infof("Client %d. Status changed to RUNNING", c.Id)
	return StateRunning
}

// Compute the size of the next chunk to read from file
func (c *FileClient) computeNextChunk() int32 {
	if c.size-c.bytesRead < defaultChunkSize {
		return c.size - c.bytesRead
	}

	return defaultChunkSize
}

// Return true if there is still data to be read from file
func (c *FileClient) hasNextChunk() bool {
	return c.size < c.bytesRead
}
