package reader

import (
	"github.com/golang/glog"
)

const (
	StateIdle = iota

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
)

type State int

// Client reads data from file and send data notification to clients.
type Client struct {
	Id int

	// cache
	cache *cache

	// state
	State State

	done chan struct{}
}

// New creates a new logger
func NewClient(id int, out chan interface{}) *Client {
	c := &Client{
		Id:    id,
		cache: newCache(),
		done:  make(chan struct{}),
		State: StateIdle,
	}

	return c
}

// Start the logger. It runs the fetcher in a go routine.
func (c *Client) Start(reader FileReader) int {
	if c.IsRunning() {
		return c.ID
	}

	glog.Infof("Starting logging with logger %d", c.ID)
	c.fetcher = newFetcher(c.ID)
	go c.fetcher.fetch(reader, c)

	return c.ID
}

// Stop stop reading the file. It doesn't disconnect the client.
// it is just stop reading the file.
func (c *Client) Stop() {
	if c.IsRunning() {
		glog.Info("Closing logger")

		c.fetcher.close()

		c.fetcher = nil
		glog.V(1).Infof("Fetcher closed. Client state: %+v", c.State)

		glog.V(1).Infof("Cached closed and cleared")
		c.cache.clear()

		close(c.done)
	}
}

// IsRunning return true if logger is running
func (c *Client) IsRunning() bool {
	return c.fetcher != nil
}

// RequestData reads `size` bytes from cache at offset `offset`.
// It returns an array of bytes and the number of bytes actual read.
func (c *Client) RequestData(offset int64, size int) ([]byte, int) {
	data := make([]byte, size)
	n, _ := c.cache.ReadAt(data, offset)

	return data, n
}

// CacheSize returns the size of the cache.
func (c *Client) CacheSize() int {
	return len(c.cache.data)
}

// WriteData writes data to cache.
func (c *Client) WriteData(data []byte) {
	// Handle new data from fetcher.
	prevSize := c.cache.size
	n, _ := c.cache.Write(data)
	if prevSize == MaxCacheSize {
		prevSize -= int64(n)
	}

	// Create a new data notification to be sent to clients
	notification := DataNotification{
		ID:           c.ID,
		Size:         c.cache.size,
		PreviousSize: prevSize,
	}
	c.out <- notification
}

// Error sends a change in state notification.
func (c *Client) Error(stderr, err error) {
	if stateChanged := c.State.HandleStateChange(stderr, err); stateChanged {
		c.out <- *(c.State)
	}
}
