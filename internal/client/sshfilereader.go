package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/ssh"
)

var (
	// ErrNofile means that the remote file doesn't exist or the user has no permission to read it.
	ErrRead = errors.New("read error")

	// ErrClient means the ssh connection is down
	ErrClient = errors.New("client error")
)

// SSHFileReader implements DataSourceReader.
// It reads from a file over a ssh connection.
type SSHFileReader struct {
	client *ssh.Client

	// filename
	filename string

	// number of bytes read from file
	bytesRead int32

	// size in bytes
	size int32
}

// NewSSHFileReader returns a SSHFileReader. The remoteClient has to be already connected.
// New will try to stat each file in the filepaths in order to create a Logfile for
// each filepath
func NewSSHFileReader(c *ssh.Client, filename string) *SSHFileReader {

	r := SSHFileReader{
		client: c,
	}

	return &r
}

// Close closes the connection
func (r *SSHFileReader) Close() {
	r.client.Close()
}

// Implementation of ReaderAt interface.
func (r *SSHFileReader) ReadAt(p []byte, off int64) (int, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := readAtCommand(r.filename, off, int32(len(p)))
	glog.V(4).Infof("\n\n ---- Running command: %s  -----", cmd)

	err := r.client.Cmd(cmd).SetStdio(&stdout, &stderr).Run()
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("%v: %w", errors.New(string(stderr.Bytes())), ErrClient)
	}

	n, err := stdout.Read(p)
	if err != nil {
		return n, fmt.Errorf("%v: %w", err, ErrRead)
	}
	return n, nil
}

// Return the size of the file and any error encountered.
func (r *SSHFileReader) Size() (int32, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := r.client.Cmd(statCommand(r.filename)).SetStdio(&stdout, &stderr).Run()
	if err != nil {
		if len(stderr.Bytes()) == 0 {
			return 0, ErrClient
		} else {
			return 0, fmt.Errorf("error: %s: %w", string(stderr.Bytes()), ErrClient)
		}
	}

	size, err := strconv.ParseInt(strings.Trim(stdout.String(), "\n"), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %w", ErrRead)
	}
	return int32(size), nil
}

/*
  Return a bash command to read a chunk from a file.
*/
func readAtCommand(filename string, offset int64, size int32) string {
	return fmt.Sprintf("tail -c+%d %s | head -c%d",
		offset,
		filename,
		size)
}

/*
StatCommand returns the command for reading total size of file
*/
func statCommand(filename string) string {
	return fmt.Sprintf("stat --format %%s %s", filename)
}
