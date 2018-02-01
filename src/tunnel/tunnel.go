package main

import (
	"fmt"
	"io/ioutil"
	"log"
	tunnel "tunnel/lib"

	"golang.org/x/crypto/ssh"

	"github.com/caarlos0/env"
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

type TunnelEnvironmentConfig struct {
	CertificatePath string `env:"CERTPATH"`
	CertificatePass string `env:"CERTPASS"`
	SSHUser         string `env:"SSHUSER"`
	SSHHost         string `env:"SSHHOST"`
	SSHPort         int    `env:"SSHPORT"`
	TargetHost      string `env:"TARGETHOST"`
	TargetPort      int    `env:"TARGETPORT"`
	SourceHost      string `env:"SOURCEHOST"`
	SourcePort      int    `env:"SOURCEPORT"`
}

func main() {

	// parse environment
	cfg := TunnelEnvironmentConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error Parsing Environment: %s", err)
		return
	}

	fmt.Printf("%+v\n", cfg)

	authMethod, err := PublicKeyAuthMethod(cfg.CertificatePath, cfg.CertificatePass)
	if err != nil {
		log.Fatalf("Error reading private key: %s", err)
		return
	}

	tunnel := &tunnel.Tunnel{
		Config: &ssh.ClientConfig{User: cfg.SSHUser, Auth: []ssh.AuthMethod{authMethod}},
		Proxy:  &tunnel.Endpoint{Host: cfg.SSHHost, Port: cfg.SSHPort},
		Source: &tunnel.Endpoint{Host: cfg.SourceHost, Port: cfg.SourcePort},
		Target: &tunnel.Endpoint{Host: cfg.TargetHost, Port: cfg.TargetPort},
	}

	if err := tunnel.Start(); err != nil {
		log.Fatalf("Error creating SSH Tunnel: %s", err)
	}
}
