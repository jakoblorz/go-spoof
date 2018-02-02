package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/caarlos0/env"
)

// SpoofEnvironmentConfig recieves the Environment Variables
type SpoofEnvironmentConfig struct {
	TargetHost string `env:"TARGETHOST"`
	TargetPort int    `env:"TARGETPORT"`
	SourcePort int    `env:"SOURCEPORT"`
	Message    string `env:"MESSAGE"`
}

func main() {

	// parse environment
	cfg := SpoofEnvironmentConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error Parsing Environment: %s", err)
		return
	}

	fmt.Printf("%+v\n", cfg)

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalf("Error Resolving Source UDP Address: %s\n", err)
		return
	}

	local := &net.UDPAddr{
		IP:   conn.LocalAddr().(*net.UDPAddr).IP,
		Port: cfg.SourcePort,
	}

	if err := conn.Close(); err != nil {
		log.Fatalf("Error Closing Source UDP Address Resolving Connection: %s", err)
		return
	}

	remote, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.TargetHost, cfg.TargetPort))
	if err != nil {
		log.Fatalf("Error Resolving Remote UDP Address: %s\n", err)
	}

	connection, err := net.DialUDP("udp", local, remote)
	if err != nil {
		log.Fatalf("Error Connecting To Remote Server: %s\n", err)
	}

	fmt.Printf("> Locale Address %+v\n", connection.LocalAddr())
	fmt.Printf("> Remote Address %+v\n", connection.RemoteAddr())

	fmt.Printf("> Writing Message \"%s\"", cfg.Message)

	wbytes, err := connection.Write([]byte(cfg.Message))
	if err != nil {
		log.Fatalf(" -> Error Writing Message: %s\n", err)
		return
	}

	fmt.Printf(" -> Wrote Total Of %d Bytes\n", wbytes)

	defer connection.Close()

	buffer := make([]byte, 1024)
	bytes, _, err := connection.ReadFromUDP(buffer)
	if err != nil {
		log.Fatalf("> Error Recieving Message: %s\n", err)
		return
	}

	fmt.Printf("> Recieved Message (%d bytes): %02X\n", bytes, buffer[0:bytes])

	msg, err := hex.DecodeString(fmt.Sprintf("%02X", buffer[0:bytes]))
	if err != nil {
		log.Fatalf("> Error Converting Byte Buffer To String: %s\n", err)
		return
	}

	fmt.Printf("%s\n", string(msg))
}
