package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/tupyy/lazylogger/internal/conf"
	"golang.org/x/crypto/ssh"
)

type remoteScriptType byte
type remoteShellType byte

const (
	cmdLine remoteScriptType = iota
	rawScript
	scriptFile

	interactiveShell remoteShellType = iota
	nonInteractiveShell
)

type Client struct {
	client *ssh.Client
}

// Dial a new connection using ConfigurationEntry.
// If jump host is defined, then it will dial a connection using jump host and create a tunnel.
func DialWithEntry(conf conf.ConfigurationEntry) (*Client, error) {
	host := conf.Host
	if len(conf.JumpHost.Address) > 0 {
		return dialWithJumpHost(host, conf.JumpHost)
	} else {
		if len(host.Key) > 0 {
			return dialWithKey(host.String(), host.Username, host.Key)
		} else {
			return dialWithPasswd(host.String(), host.Username, host.Password)
		}
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

// Cmd create a command on client
func (c *Client) Cmd(cmd string) *remoteScript {
	return &remoteScript{
		_type:  cmdLine,
		client: c.client,
		script: bytes.NewBufferString(cmd + "\n"),
	}
}

func (c *Client) NewSession() (*ssh.Session, error) {
	return c.client.NewSession()
}

func (c *Client) IsConnected() bool {
	return c.IsConnected()
}

// dial0WithPasswd starts a client connection to the given SSH server with passwd authmethod.
func dialWithPasswd(addr, user, passwd string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         5 * time.Second,
	}

	return dial0("tcp", addr, config)
}

// dial0WithKey starts a client connection to the given SSH server with key authmethod.
func dialWithKey(addr, user, keyfile string) (*Client, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         5 * time.Second,
	}

	return dial0("tcp", addr, config)
}

// dial0WithKeyWithPassphrase same as dial0WithKey but with a passphrase to decrypt the private key
func dialWithKeyWithPassphrase(addr, user, keyfile string, passphrase string) (*Client, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         5 * time.Second,
	}

	return dial0("tcp", addr, config)
}

func dialWithJumpHost(jumpHost, host conf.Host) (*Client, error) {

	var auth ssh.AuthMethod
	var err error
	if len(jumpHost.Key) > 0 {
		auth, err = privateKeyFile(jumpHost.Key)
		if err != nil {
			return nil, err
		}
	} else {
		auth = ssh.Password(jumpHost.Password)
	}

	jumpHostConfig := &ssh.ClientConfig{
		User:            jumpHost.Username,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Timeout:         5 * time.Second,
	}

	jumpConn, err := ssh.Dial("tcp", jumpHost.String(), jumpHostConfig)
	if err != nil {
		return nil, fmt.Errorf("dial error to jump host: %w", err)
	}

	remoteConn, err := jumpConn.Dial("tcp", host.String())
	if err != nil {
		return nil, fmt.Errorf("dial error from jumphost to remote: %w", err)
	}

	if len(host.Key) > 0 {
		auth, err = privateKeyFile(host.Key)
		if err != nil {
			return nil, err
		}
	} else {
		auth = ssh.Password(host.Password)
	}

	remoteHostConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Timeout:         5 * time.Second,
	}

	c, chans, reqs, err := ssh.NewClientConn(remoteConn, host.String(), remoteHostConfig)
	if err != nil {
		return nil, fmt.Errorf("create ssh client error: %w", err)
	}

	return &Client{ssh.NewClient(c, chans, reqs)}, nil
}

func privateKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

// dial0 starts a client connection to the given SSH server.
// This is wrap the ssh.dial0
func dial0(network, addr string, config *ssh.ClientConfig) (*Client, error) {
	client, err := ssh.Dial(network, addr, config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

type remoteScript struct {
	client     *ssh.Client
	_type      remoteScriptType
	script     *bytes.Buffer
	scriptFile string
	err        error

	stdout io.Writer
	stderr io.Writer
}

// Run
func (rs *remoteScript) Run() error {
	if rs.err != nil {
		fmt.Println(rs.err)
		return rs.err
	}

	if rs._type == cmdLine {
		return rs.runCmds()
	} else {
		return errors.New("Not supported remoteScript type")
	}
}

func (rs *remoteScript) Output() ([]byte, error) {
	if rs.stdout != nil {
		return nil, errors.New("Stdout already set")
	}
	var out bytes.Buffer
	rs.stdout = &out
	err := rs.Run()
	return out.Bytes(), err
}

func (rs *remoteScript) Cmd(cmd string) *remoteScript {
	_, err := rs.script.WriteString(cmd + "\n")
	if err != nil {
		rs.err = err
	}
	return rs
}

func (rs *remoteScript) SetStdio(stdout, stderr io.Writer) *remoteScript {
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}

func (rs *remoteScript) runCmd(cmd string) error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = rs.stdout
	session.Stderr = rs.stderr

	if err := session.Run(cmd); err != nil {
		return err
	}
	return nil
}

func (rs *remoteScript) runCmds() error {
	for {
		statment, err := rs.script.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := rs.runCmd(statment); err != nil {
			return err
		}
	}

	return nil
}

type remoteShell struct {
	client     *ssh.Client
	requestPty bool

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (rs *remoteShell) SetStdio(stdin io.Reader, stdout, stderr io.Writer) *remoteShell {
	rs.stdin = stdin
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}
