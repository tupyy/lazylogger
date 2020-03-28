package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/ssh"
)

var (
	// DefaultChunkSize set to 4K
	DefaultChunkSize = int32(40 * 1024)
)

// ErrNofile means that the remote file doesn't exist or the user has no permission to read it.
var ErrNofile = errors.New("file don't exist")

// ErrClient means the ssh connection is down
var ErrClient = errors.New("client error")

// ErrInvalidSize means that the size as string returned by the stat command
// cannot be parsed into an uint32
var ErrInvalidSize = errors.New("invalid size")

/*
logFile keeps the information about the file to be logged
*/
type logFile struct {
	Path string

	// number of bytes read from file
	BytesRead int32

	// current skip value
	Skip int32

	// size in bytes
	Size int32
}

/*
NextChunkCommand return the dd command to be executed in the shell
in order to get the next chuck of data
*/
func (log *logFile) NextChunkCommand(chunkSize int32) string {
	return fmt.Sprintf("tail -c+%d %s | head -c%d",
		log.BytesRead,
		log.Path,
		chunkSize)

}

/*
StatCommand returns the command for reading total size of file
*/
func (log *logFile) StatCommand() string {
	return fmt.Sprintf("stat --format %%s %s", log.Path)
}

// RemoteReader keeps track of file to be logged on a remote host.
// Only one file can be watched at the time.
type RemoteReader struct {
	client *ssh.Client
	file   logFile
}

// NewRemoteReader returns a RemoteReader. The remoteClient has to be already connected.
// New will try to stat each file in the filepaths in order to create a Logfile for
// each filepath
func NewRemoteReader(c *ssh.Client, file string) *RemoteReader {

	r := RemoteReader{
		client: c,
		file:   logFile{Path: file, BytesRead: 0, Skip: 0, Size: 0},
	}

	return &r
}

// Close closes the connection
func (r *RemoteReader) Close() {
	r.client.Close()
}

// ReadNextChunk reads the next chunk from file.
func (r *RemoteReader) ReadNextChunk() ([]byte, error, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	// protect the size. We could set the new size at the same time
	size := r.file.Size

	cmd := r.file.NextChunkCommand(computeNextChunkSize(size, r.file.BytesRead, DefaultChunkSize))
	glog.V(4).Infof("\n\n ---- Running command: %s  -----", cmd)

	err := r.client.Cmd(cmd).SetStdio(&stdout, &stderr).Run()
	if err != nil && err != io.EOF {
		return []byte{}, errors.New(string(stderr.Bytes())), err
	}

	bytesRead := uint32(stdout.Len())

	if bytesRead > 0 {
		r.file.Skip++
		r.file.BytesRead += int32(stdout.Len())
		return stdout.Bytes(), nil, nil
	}

	return []byte{}, nil, nil

}

// HasNextChunk returns true if there is more data to be read from file.
// It does not update the size of the file
func (r *RemoteReader) HasNextChunk() bool {
	return r.file.Size > r.file.BytesRead
}

// ReadChunk reads a single chunk at offset offset.
// Return EOF if offset >= size
func (r *RemoteReader) ReadChunk(offset int) (string, error) {
	return "chunk", nil
}

// Rewind set bytesRead to zero
func (r *RemoteReader) Rewind() {
	r.file.BytesRead = 0
	r.file.Size = 0
}

// GetSize returns the size of the file
func (r *RemoteReader) GetSize() int32 {
	return r.file.Size
}

// SetSize set file size
func (r *RemoteReader) SetSize(size int32) {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	r.file.Size = size

	// if size > MaxCacheSize and bytesRead == 0, skip the beggining of the file and set bytesRead to MaxCacheSize
	if size > MaxCacheSize && r.file.BytesRead == 0 {
		r.file.BytesRead = size - MaxCacheSize
	}

	mutex.Unlock()
}

// FetchSize will fetch the size from the remote client
// FetchSize returns two errors: the first one is when something is wrong with the file but the connection is ok, the second when the client is down.
func (r *RemoteReader) FetchSize() (int32, error, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := r.client.Cmd(r.file.StatCommand()).SetStdio(&stdout, &stderr).Run()
	if err != nil {
		if len(stderr.Bytes()) == 0 {
			return 0, nil, ErrClient
		} else {
			return 0, errors.New(string(stderr.Bytes())), nil
		}
	}

	size, err := strconv.ParseInt(strings.Trim(stdout.String(), "\n"), 10, 32)
	if err != nil {
		return 0, ErrInvalidSize, nil
	}
	return int32(size), nil, nil
}

// computeNextChunkSize compute the size of the next chunk in bytes
func computeNextChunkSize(size, totalBytesRead, DefaultChunkSize int32) int32 {
	if size == totalBytesRead {
		return 0
	}
	chunk := DefaultChunkSize
	if totalBytesRead+chunk > size {
		return size - totalBytesRead
	}

	return chunk
}
