package ssh

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/golang/glog"
	"github.com/tupyy/lazylogger/internal/conf"
)

type SSHPool struct {
	clients map[string]*Client
}

func NewSSHPool() *SSHPool {
	return &SSHPool{clients: make(map[string]*Client)}
}

// Returns a hash created from Host as string.
func createHash(addr, user, pwd string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%s%s", addr, user, pwd)))
	return string(fmt.Sprintf("%x", hash.Sum(nil)))
}

// Connect looks for a existing client in clients. If a client is found, returns it.
// If not, it dials a new connection.
// Connect doesn't check if a connection is still alive
func (sshPool *SSHPool) Connect(conf conf.LoggerConfiguration) (*Client, error) {
	host := conf.Host
	hashID := createHash(host.String(), host.Username, host.Password)
	v, ok := sshPool.clients[hashID]
	if !ok {
		glog.Infof("Connection exists for %s with user %s:", host.String(), host.Username)
		if len(conf.JumpHost.Address) > 0 {
			return sshPool.connectWithJumpHost(host, conf.JumpHost)
		} else {
			if len(host.Key) > 0 {
				return sshPool.connectToHostWithKey(host.String(), host.Username, host.Key)
			} else {
				return sshPool.connectToHost(host.String(), host.Username, host.Password)
			}
		}
	}

	// if the connection is not alive, try to reconnect
	if !isAlive(v) {
		if len(host.Key) > 0 {
			return sshPool.connectToHostWithKey(host.String(), host.Username, host.Key)
		} else {
			return sshPool.connectToHost(host.String(), host.Username, host.Password)
		}
	}

	return v, nil
}

func (sshPool *SSHPool) Disconnect() {
	for _, c := range sshPool.clients {
		c.Close()
	}
}

// dial the connection and save the client to clients
func (sshPool *SSHPool) connectToHostWithKey(addr, user, key string) (*Client, error) {
	hashID := createHash(addr, user, key)
	client, err := DialWithKey(addr, user, key)
	if err != nil {
		return nil, err
	}
	sshPool.clients[hashID] = client

	glog.Infof("Connected to: %s with user: %s", addr, user)
	return client, nil
}

// dial the connection and save the client to clients
func (sshPool *SSHPool) connectToHost(addr, user, pwd string) (*Client, error) {
	hashID := createHash(addr, user, pwd)
	client, err := DialWithPasswd(addr, user, pwd)
	if err != nil {
		return nil, err
	}
	sshPool.clients[hashID] = client

	glog.Infof("Connected to: %s with user: %s", addr, user)
	return client, nil
}

// dial the connection and save the client to clients
func (sshPool *SSHPool) connectWithJumpHost(host, jumpHost conf.Host) (*Client, error) {
	hashID := createHash(host.String(), host.Username, host.Key)
	client, err := DialWithJumpHost(jumpHost, host)
	if err != nil {
		return nil, err
	}
	sshPool.clients[hashID] = client

	glog.Infof("Connected to: %s with user: %s", host.String(), host.Username)
	return client, nil
}

func isAlive(client *Client) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := client.Cmd("echo stdout").Cmd(">&2 echo stderr").SetStdio(&stdout, &stderr).Run()
	if err != nil {
		return false
	}

	return true
}
