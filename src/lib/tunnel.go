package spoof

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// Credits @svett:
// https://gist.github.com/svett/5d695dcc4cc6ad5dd275

// Credits @josephspurrier
// https://gist.github.com/josephspurrier/e83bcdbf9e6865500004

// PublicKeyAuthMethod creates a Public Key as ssh.AuthMethod
// from a private key file
func PublicKeyAuthMethod(file string) (ssh.AuthMethod, error) {
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

// ParseAddress splits the address into its parts and extracts
// the hostname and the port
func ParseAddress(address string) (string, int64, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", 0, err
	}

	host, _port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.ParseInt(_port, 10, 0)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

// Endpoint represents a single Server
// with Hostname and Port
type Endpoint struct {
	Host string
	Port int64
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// Tunnel represents the tunneling
// components: source, proxy and target
// plus the config for the proxy
type Tunnel struct {
	Source *Endpoint
	Proxy  *Endpoint
	Target *Endpoint
	Config *ssh.ClientConfig
}

// Start starts a listener on the Source Server. Once connected it spawns
// a forwarding session (Forward())
func (tunnel *Tunnel) Start() error {
	listener, err := net.Listen("udp", tunnel.Source.String())
	if err != nil {
		fmt.Printf("Could not connect to Source Server %s\n", err)
		return err
	}

	defer listener.Close()

	return tunnel.StartFromListener(listener)
}

// StartFromConnection starts a forwarding session (Forward()) right
// from an existing Connection
func (tunnel *Tunnel) StartFromConnection(connection net.Conn) error {
	go tunnel.Forward(connection)

	return nil
}

// StartFromListener starts a forwarding session (Forward()) right
// when a connection is established
func (tunnel *Tunnel) StartFromListener(listener net.Listener) error {

	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}

		tunnel.StartFromConnection(connection)
	}
}

// Forward connectes to the SSH Server, then connecting
// to the Target Server
func (tunnel *Tunnel) Forward(conn net.Conn) {
	sshconn, err := ssh.Dial("tcp", tunnel.Proxy.String(), tunnel.Config)
	if err != nil {
		fmt.Printf("Could not connect to SSH-Proxy Server: %s\n", err)
		return
	}

	connection, err := sshconn.Dial("udp", tunnel.Target.String())
	if err != nil {
		fmt.Printf("Could not connect to Target Server %s\n", err)
	}

	copy := func(write, read net.Conn) {
		_, err := io.Copy(write, read)
		if err != nil {
			fmt.Printf("Connection Copy Error: %s\n", err)
			return
		}
	}

	go copy(conn, connection)
	go copy(connection, conn)
}
