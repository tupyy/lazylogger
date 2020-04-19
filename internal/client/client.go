package client

import (
	"errors"
	"io"

	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/datasource"
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

// The implementations of this interface must provide a way to read data from a source.
type Datasource interface {
	io.ReaderAt
	io.Closer

	// Size return the size of the underlining datasource
	Size() (int32, error)
}

// Client implements basic mechanism to read from a data source.
//
// This mechanism has two actions fetching size and fetching data from data source.
// Client store the size of the data source internally and regulary checks if the fetched size is greater than the stored size.
// If true, it starts fetching data between the old size and the new size otherwise it will wait 1 sec and tries again.
//
// Fetching data is done in chunks. The max size of a chunk is 100kb. After each fetched chunk, it computes the size of the next chunk which is
// the difference between the size of the data source and the total bytes read from source.
// The next chunk cannot be greater than 100kb.
// When all the data is fetched from source, the client starts to fetch the size again.
// The data fetched from data source is kept in cache of 300kb in size. The cache rotates meaning il will overwrite the old data if size reached the maximum size.
//
// There are two types of errors which can be returned by a data source: ErrDatasource and ErrRead.
// ErrDatasource means that there is a fatal error with the data source like the lost of ssh connection.
// In the event of a ErrDatasource, the client stops and state of the client becomes StateError
//
// ErrRead is a reading error. Although the datasource is fine there is a problem reading the underlying source. For example, for SshDatasource, it can means that
// the file is no longer available. In this case, the clients pass to StateDegrated but keeps trying to fetch size. If at some point, the returned error is nil,
// the state changes to StateRunning and the process starts again.
//
// When client is stopped, it stops the go routine which implements the fetching mechanism but it does not close the data source.
//
// Writers can be registers to client. A writer is any type which implements the io.Writer interface. When new data is fetched, the client automatically writes
// the data to any registered writer.
// Client implements the io.Reader and io.ReaderAt interfaces meaning anyone can read from a client using Read and ReadAt methods. In this case, the data is read
// from the cache. The implementations of ReaderAt do not block if data is not available and it return the number of bytes available if any and io.EOF error.
type Client struct {

	// id of the client
	Id int

	// State of the client
	State State

	// datasource
	Datasource Datasource

	done chan struct{}

	// cache
	cache *cache

	// data source reader

	// bytes read
	bytesRead int32

	// size represents the total bytes read from the data source. It can be greater the size of the cache.
	size int32

	// list of writers
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
func NewFileClient(id int, reader Datasource) *Client {
	c := &Client{
		Id:         id,
		cache:      newCache(),
		done:       make(chan struct{}),
		writers:    []io.Writer{},
		Datasource: reader,
	}

	return c
}

// Start the client
func (c *Client) Start() {
	c.State = StateRunning
	go c.fetch()
}

// Stop the fetch go routine.
// It does nothing if fetch is already stopped.
func (c *Client) Stop() {
	if c.State != StateStopped {
		c.done <- struct{}{}
		c.State = StateStopped
	}
}

// Implementation of ReaderAt interface reading from cache.
func (c *Client) ReadAt(p []byte, off int64) (n int, err error) {
	return c.cache.ReadAt(p, off)
}

// Implementation of Reader interface reading from cache.
func (c *Client) Read(p []byte) (n int, err error) {
	return c.cache.ReadAt(p, 0)
}

// Size returns the size of the cache.
func (c *Client) Size() int32 {
	return int32(len(c.cache.data))
}

func (c *Client) AddWriter(w io.Writer) {
	c.writers = append(c.writers, w)
}

// Fetch the data from reader
// It starts by fetching the size of data source. If the fetched size is greater than the old one, it will start fetching data.
// When all the data has been fetched, it starts fetching size again. So on until it stops.
func (c *Client) fetch() {
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
			c.State = StateStopped
			return
		case <-fetchData:
			chunk := c.computeNextChunk()
			glog.V(2).Infof("Fetching next chunk of %d bytes.", chunk)

			p := make([]byte, chunk)
			bytesRead, err := c.Datasource.ReadAt(p, int64(c.bytesRead))
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
			newSize, err := c.Datasource.Size()
			if err != nil {
				c.State = c.handleStateChange(err)
				if c.State == StateError {
					go stop()
				} else {
					go startFetchingSize()
				}
			} else {
				if c.size < newSize {
					c.size = newSize
					go startFetchingData()
				} else {
					go startFetchingSize()
				}
			}
		}
	}
}

// Error of type ErrClient change the state to ERROR because they must represents error in fatal error in readers.
// Usually, this means that the data source has crashed (e.g. ssh connection ended).
// Error of type ErrRead change state to DEGRADED meaning that the reading operation failed but the connection is still ok.
func (c *Client) handleStateChange(err error) State {
	glog.V(2).Infof("%s", err)
	if errors.Is(err, datasource.ErrDatasource) {
		glog.V(1).Infof("Client: %d. Status changed to ERROR", c.Id)
		return StateError
	} else if errors.Is(err, datasource.ErrRead) {
		glog.V(1).Infof("Client %d. Status changed to DEGRADED", c.Id)
		return StateDegrated
	}
	glog.V(1).Infof("Client %d. Status changed to RUNNING", c.Id)
	return StateRunning
}

// Compute the size of the next chunk to read from file
func (c *Client) computeNextChunk() int32 {
	if c.size-c.bytesRead < defaultChunkSize {
		return c.size - c.bytesRead
	}

	return defaultChunkSize
}

// Return true if there is still data to be read from file
func (c *Client) hasNextChunk() bool {
	return c.size < c.bytesRead
}
