package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"tunnel/lib"

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
	log.SetFlags(log.Lshortfile)

	var SSHKEY string
	if key, exists := os.LookupEnv("SSHKEY"); exists {
		SSHKEY = key
	} else {
		fmt.Printf("SSH Key not specified (NO SSHKEY); falling back to default ~/.ssh/is_rsa\n")
		SSHKEY = "~/.ssh/id_rsa.pub"
	}

	var SSHUSER string
	if user, exists := os.LookupEnv("SSHUSER"); exists {
		SSHUSER = user
	} else {
		log.Fatal("SSH User not specifed (NO SSHUSER); cannot fall back to default: no default\n")
		return
	}

	var SSHPASS string
	if pass, exists := os.LookupEnv("SSHPASS"); exists {
		SSHPASS = pass
	} else {
		log.Fatalf("SSH Key Password not specified (NO SSHPASS); cannot fall back to default: no default")
		return
	}

	var SSHAUTH ssh.AuthMethod
	if auth, err := PublicKeyAuthMethod(SSHKEY, SSHPASS); err == nil {
		SSHAUTH = auth
	} else {
		log.Fatalf("SSH Key could not be read: %s\n", err)
		return
	}

	config := &ssh.ClientConfig{
		User: SSHUSER,
		Auth: []ssh.AuthMethod{SSHAUTH},
	}

	var REMOTE string
	if remote, exists := os.LookupEnv("REMOTE"); exists {
		REMOTE = remote
	} else {
		log.Fatalf("REMOTE Address not specified (NO REMOTE); cannot fall back to default: no default\n")
		return
	}

	HREMOTE, PREMOTE, err := ParseAddress(REMOTE)
	if err != nil {
		log.Fatalf("Error when parsing REMOTE Address (ERR REMOTE): %s\n", err)
		return
	}

	var PROXY string
	if proxy, exists := os.LookupEnv("PROXY"); exists {
		PROXY = proxy
	} else {
		log.Fatalf("PROXY Address not specified (NO PROXY); cannot fall back to default: no default\n")
		return
	}

	HPROXY, PPROXY, err := ParseAddress(PROXY)
	if err != nil {
		log.Fatalf("Error when parsing PROXY Address (ERR PROXY): %s\n", err)
	}

	var LOCAL string
	if local, exists := os.LookupEnv("LOCAL"); exists {
		LOCAL = local
	} else {
		log.Fatalf("LOCAL Address not specified (NO LOCAL); cannot fall back to default: no default\n")
		return
	}

	HLOCAL, PLOCAL, err := ParseAddress(LOCAL)
	if err != nil {
		log.Fatalf("Error when parsing LOCAL Address (ERR LOCAL): %s\n", err)
		return
	}

	pspoof, exists := os.LookupEnv("SPOOF_PORT")
	if !exists {
		log.Fatalf("SPOOF_PORT is not specified (NO SPOOF_PORT); cannot fall back to default: no default\n")
		return
	}

	PSPOOF, err := strconv.Atoi(pspoof)
	if err != nil {
		log.Fatalf("SPOOF_PORT could not be converted to int (ERR SPOOF_PORT): %s\n", err)
		return
	}

	var CSPOOF string
	if cspoof, exists := os.LookupEnv("SPOOF_CONTENT"); exists {
		CSPOOF = cspoof
	} else {
		log.Fatalf("SPOOF_CONTENT is not specified (NO SPOOF_CONTENT); cannot fall back to default: no default\n")
		return
	}

	fmt.Println("Starting Tunnel: ")
	fmt.Println(" - $SSHUSER=" + SSHUSER)
	fmt.Println(" - $SSHKEY=" + SSHKEY)
	fmt.Println("   -> Create Local Tunnel Endpoint: ")
	fmt.Printf("      - $LOCALPORT=%d\n", PLOCAL)
	fmt.Println("      - $LOCALHOST=" + HLOCAL)
	fmt.Println("   -> Connect To SSH Tunnel Server: ")
	fmt.Printf("      - $SSHPORT=%d\n", PPROXY)
	fmt.Println("      - $SSHHOST=" + HPROXY)
	fmt.Println("   -> Connect To Remote Server: ")
	fmt.Printf("      - $REMOTEPORT=%d\n", PREMOTE)
	fmt.Println("      - $REMOTEHOST=" + HREMOTE)

	tunnel := &tunnel.Tunnel{
		Network: "udp",
		Config:  config,
		Proxy:   &tunnel.Endpoint{Host: HPROXY, Port: PPROXY},
		Source:  &tunnel.Endpoint{Host: HLOCAL, Port: PLOCAL},
		Target:  &tunnel.Endpoint{Host: HREMOTE, Port: PREMOTE},
	}

	if err := tunnel.Start(); err != nil {
		log.Fatalf("Error when creating Tunnel: %s\n", err)
		return
	}

	remoteAddr, err := net.ResolveUDPAddr(tunnel.Network, tunnel.Source.Host)
	if err != nil {
		log.Fatalf("Error when resolving Local Address: %s\n", err)
	}

	for {

		conn, err := net.DialUDP(tunnel.Network, &net.UDPAddr{
			IP:   GetOwnIP(),
			Port: PSPOOF,
		}, remoteAddr)

		if err != nil {
			log.Fatalf("Error connecting to Local Tunnel Endpoint: %s\n", err)
			return
		}

		conn.Write([]byte(CSPOOF))

		if message, err := bufio.NewReader(conn).ReadString('\n'); err != nil {
			log.Fatalf("Error recieving Response: %s", err)
		} else {
			fmt.Printf("recieved response: %s\n", message)
			conn.Close()
		}
	}
}
