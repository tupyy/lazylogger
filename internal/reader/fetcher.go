package reader

import (
	"github.com/golang/glog"
	"sync"
	"time"
)

// DataWriter is an interface that writes data fetched.
type DataWriter interface {
	WriteData(data []byte)
	Error(stderr, err error)
}

// FileReader reads data from file in small chuncks. If the size of the file has increased
// from the last read HasNextChunk will return true and ReadNextChunk reads the file until
// HasNextChunk returns false.
type FileReader interface {

	// GetSize return the size of the file.
	GetSize() int32

	// Setsize sets the size.
	// DEPRECATED. TO BE REMOVED
	SetSize(int32)

	// ReadNextChunk reads chunks until HasNextChunk returns false.
	ReadNextChunk() ([]byte, error, error)

	// HasNextChunk return true if the size fetched is bigger than the current size set by SetSize.
	HasNextChunk() bool

	// Close the file.
	Close()

	// FetchSize fetch the size of the file. It returns size of the file, stderr if the, for some reason
	// (e.g file doesn't exists anymore) size cannot be read, err if there are problems with the connection.
	FetchSize() (int32, error, error)

	// Rewind resets the size to zero. It is called if the fetched size is smaller than the current size.
	Rewind()
}

// Fetcher take care of fetching the size and data from remote host.
// It uses a non-blocking loop to fetch both size and data.
// The size of the file is fetched every 1 seconds. If the fetched size is greated than the
// last size return by the LogFile struct then fetchData is running.
// fetchData will fetch the data between fetchedSize - remotelogger.GetSize().
type fetcher struct {
	id int

	// channel to close the fetcher. It returns a channel to wait for the fetcher to close.
	closing chan chan struct{}

	// error to notify logger about error from fetching size or data
	errorCh chan struct{}

	// data is sent through this channel
	data chan []byte

	// stderr
	stderr error

	// sshConnectionErr means the connection failed
	sshConnectionErr error
}

func newFetcher(id int) *fetcher {
	return &fetcher{
		id:      id,
		closing: make(chan chan struct{}),
		errorCh: make(chan struct{}),
		data:    make(chan []byte),
	}

}

// Result of fetching size
type fetchedSizeResult struct {
	size             int32
	stderr           error
	sshConnectionErr error
}

// Result of fetching data
type fetchedDataResult struct {
	stdout           []byte // data received from remote stdout
	stderr           error
	sshConnectionErr error
}

// Asks the fetchData loop to exits and waits for a response
func (f *fetcher) close() {
	glog.Infof("Close the fetcher: %d", f.id)
	doneCh := make(chan struct{}, 1)
	f.closing <- doneCh
}

// fetch the data from the log
func (f *fetcher) fetch(fr FileReader, dataWriter DataWriter) {
	var fetchSizeDone chan fetchedSizeResult // if non-nil fetchSize is running
	var fetchDataDone chan fetchedDataResult // if non-nil fetchData is running
	var startFetchSize <-chan time.Time
	var wg sync.WaitGroup

	for {
		if fetchSizeDone == nil {
			startFetchSize = time.After(1 * time.Second)
		}

		select {
		case doneCh := <-f.closing:
			glog.V(3).Infof("Fetcher %d closed.", f.id)

			// wait for fetch size or fetch data go routine to exit
			wg.Wait()

			doneCh <- struct{}{}
			return

		case <-startFetchSize:
			glog.V(3).Infof("Fetcher: %d. Fetching size.", f.id)
			fetchSizeDone = make(chan fetchedSizeResult, 1)

			wg.Add(1)
			go func() {
				defer wg.Done()
				size, stderr, err := fr.FetchSize()
				fetchSizeDone <- fetchedSizeResult{size, stderr, err}
			}()

		case fetchedSize := <-fetchSizeDone:
			glog.V(3).Infof("Fetcher %d. Size fetched: %+v", f.id, fetchedSize)

			if fetchedSize.sshConnectionErr != nil || fetchedSize.stderr != nil {
				dataWriter.Error(fetchedSize.stderr, fetchedSize.sshConnectionErr)
			} else {

				// clean up any previous errors
				dataWriter.Error(nil, nil)

				if fetchedSize.size != fr.GetSize() {
					if fetchedSize.size < fr.GetSize() {
						// something happen to the file. instead of append
						// the file either has been rewrited or some parts have been deleted.
						// In this case, just rewind the file and start all over
						glog.V(2).Infof("Fetched size smaller than actual size.Rewind the file")
						fr.Rewind()
					}
					glog.V(3).Infof("Fetcher %d. Fetching new data of %d bytes.", f.id, fetchedSize.size-fr.GetSize())
					fr.SetSize(fetchedSize.size)
					fetchDataDone = make(chan fetchedDataResult, 1)

					wg.Add(1)
					go func() {
						defer wg.Done()
						stdout, stderr, err := f.fetchData(fr)
						fetchDataDone <- fetchedDataResult{stdout, stderr, err}
					}()
				}
			}
			fetchSizeDone = nil
		case fetchedData := <-fetchDataDone:
			glog.V(3).Infof("Fetcher %d. Data fetched: %+v.", f.id, fetchedData)
			if fetchedData.sshConnectionErr != nil || fetchedData.stderr != nil {
				dataWriter.Error(fetchedData.stderr, fetchedData.sshConnectionErr)
			} else {
				dataWriter.WriteData(fetchedData.stdout)
			}
			fetchSizeDone = nil
		}
	}
}

// fetch data from source. it reads the data until the HasNextChunk returns true.
func (f *fetcher) fetchData(fr FileReader) ([]byte, error, error) {
	var buffer = []byte{}
	for {
		if ok := fr.HasNextChunk(); ok {
			glog.V(4).Infof("Fetcher %d. Reading next chunk..", f.id)

			stdout, stderr, err := fr.ReadNextChunk()
			glog.V(4).Infof("Data read %s", string(stdout))

			if stderr != nil || err != nil {
				return []byte{}, stderr, err
			}
			buffer = append(buffer, stdout...)
		} else {
			return buffer, nil, nil
		}
	}
}
