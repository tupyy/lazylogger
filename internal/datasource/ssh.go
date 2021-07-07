package datasource

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/conf"
	"github.com/tupyy/lazylogger/internal/ssh"
)

// SSHDataSource implements DataSourceReader.
// It reads from a file over a ssh connection.
type SSHDataSource struct {

	// ssh client used to connect to remote host
	client *ssh.Client

	// connection definition
	Conf conf.ConfigurationEntry

	// File
	File string

	// number of bytes read from file
	bytesRead int32

	// size in bytes
	size int32
}

// NewSshDatasource returns a SSHDataSource. The remoteClient has to be already connected.
// New will try to stat each file in the filepaths in order to create a Logfile for
// each filepath
func NewSshDatasource(conf conf.ConfigurationEntry, file string) *SSHDataSource {
	return &SSHDataSource{
		File: file,
		Conf: conf,
	}
}

// Connect tries to dial a ssh connect to host.
// If the client is already connected, it disconnects the client before dialing a new connection.
func (d *SSHDataSource) Connect() error {
	if d.client.IsConnected() {
		d.client.Close()
		d.client = nil
	}

	client, err := ssh.DialWithEntry(d.Conf)
	if err != nil {
		return err
	}

	d.client = client
	return nil
}

// IsConnected return true if the ssh client is connected to host.
func (d *SSHDataSource) IsConnected() bool {
	return d.client.IsConnected()
}

// Close closes the connection
func (d *SSHDataSource) Close() {
	d.client.Close()
}

// Implementation of ReaderAt interface.
func (d *SSHDataSource) ReadAt(p []byte, off int64) (int, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := readAtCommand(d.File, off, int32(len(p)))
	glog.V(4).Infof("\n\n ---- Running command: %s  -----", cmd)

	err := d.client.Cmd(cmd).SetStdio(&stdout, &stderr).Run()
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("%v: %w", errors.New(string(stderr.Bytes())), ErrDatasource)
	}

	n, err := stdout.Read(p)
	if err != nil {
		return n, fmt.Errorf("%v: %w", err, ErrRead)
	}
	return n, nil
}

// Return the size of the file and any error encountered.
func (d *SSHDataSource) Size() (int32, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := d.client.Cmd(statCommand(d.File)).SetStdio(&stdout, &stderr).Run()
	if err != nil {
		if len(stderr.Bytes()) == 0 {
			return 0, ErrDatasource
		} else {
			return 0, fmt.Errorf("error: %s: %w", string(stderr.Bytes()), ErrDatasource)
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
