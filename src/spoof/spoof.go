package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// PublicKeyAuthMethod creates a Public Key as ssh.AuthMethod
// from a private key file
func PublicKeyAuthMethod(file string, password string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}

// GetOwnIP gets own IP Address
func GetOwnIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
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

func main() {
	log.Printf("Spoofing will soon be starting\n")
}
