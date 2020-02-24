package docker

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerReader reads docker containers' logs.
// It implements the FileReader interface.
type DockerReader struct {

	// docker client
	client *client.Client

	// id of the container which logs are read
	containerId string
}

func NewDockerLogReader(host, version string) (*DockerReader, error) {
	var defaultTtransport http.RoundTripper = &http.Transport{Proxy: nil}
	c := &http.Client{Transport: defaultTtransport}

	if host == "" {
		host = client.DefaultDockerHost
	}

	if version == "" {
		version = client.DefaultVersion
	}

	cli, err := client.NewClient(host, version, c, nil)
	if err != nil {
		return &DockerReader{}, err
	}

	return &DockerReader{client: cli}, nil
}

// Clone returns a clone of DockerReader.
func (d *DockerReader) Clone() *DockerReader {
	var clone = DockerReader{client: d.client}
	return &clone
}

//ListContainers is a wrapper around ContainerList with timeout context of 5 seconds.
func (d *DockerReader) ListContainers() ([]types.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.client.ContainerList(ctx, types.ContainerListOptions{})
}

// Reads the logs from containerId and return an array of bytes, number of bytes read and error if any.
func containerLogs(client *client.Client, containerId string) ([]byte, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reader, err := client.ContainerLogs(ctx, "container_id", types.ContainerLogsOptions{})
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, reader)
	if err != nil && err != io.EOF {
		return []byte{}, n, err
	}

	return buf.Bytes(), n, nil

}
