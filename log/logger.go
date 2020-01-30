package log

import (
	"github.com/golang/glog"
)

type Logger struct {
	ID int

	// Outbound channel. Clients reading from this channel can read DataNotification and state messages.
	out chan interface{}

	//fetches the data from file
	fetcher *fetcher

	// cache
	cache *cache

	// state
	State *State

	done chan struct{}
}

// New creates a new logger and return its ID
func NewLogger(id int, out chan interface{}) *Logger {
	l := &Logger{
		ID:      id,
		out:     out,
		fetcher: nil,
		cache:   newCache(),
		done:    make(chan struct{}),
		State:   NewState(id),
	}

	return l
}

// Start starts reading the file.
func (l *Logger) Start(reader FileReader) int {
	if l.IsRunning() {
		return l.ID
	}

	glog.Infof("Starting logging with logger %d", l.ID)
	l.fetcher = newFetcher(l.ID)
	go l.fetcher.fetch(reader, l)

	return l.ID
}

// Stop stop reading the file. It doesn't disconnect the client.
// it is just stop reading the file.
func (l *Logger) Stop() {
	if l.IsRunning() {
		glog.Info("Closing logger")

		l.fetcher.close()

		l.fetcher = nil
		glog.V(1).Infof("Fetcher closed. Logger state: %+v", l.State)

		glog.V(1).Infof("Cached closed and cleared")
		l.cache.clear()

		close(l.done)
	}
}

// IsRunning return true is logger is running
func (l *Logger) IsRunning() bool {
	return l.fetcher != nil
}

// RequestData reads `size` bytes from cache at offset `offset`.
// It returns an array of bytes and the number of bytes actual read.
func (l *Logger) RequestData(offset int64, size int) ([]byte, int) {
	data := make([]byte, size)
	n, _ := l.cache.ReadAt(data, offset)

	return data, n
}

func (l *Logger) CacheSize() int {
	return len(l.cache.data)
}

func (l *Logger) WriteData(data []byte) {
	// Handle new data from fetcher.
	prevSize := l.cache.size
	n, _ := l.cache.Write(data)
	if prevSize == MaxCacheSize {
		prevSize -= int64(n)
	}

	// Create a new data notification to be sent to clients
	notification := DataNotification{
		ID:           l.ID,
		Size:         l.cache.size,
		PreviousSize: prevSize,
	}
	l.out <- notification
}

func (l *Logger) Error(stderr, err error) {
	if stateChanged := l.State.HandleStateChange(stderr, err); stateChanged {
		l.out <- *(l.State)
	}
}
