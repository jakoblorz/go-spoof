package main

import (
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
	CertificatePath string `ENV:"CERTPATH"`
	CerfiticatePass string `ENV:"CERTPASS"`
	SSHUser         string `ENV:"SSHUSER"`
	SSHHost         string `ENV:"SSHHOST"`
	SSHPort         int    `ENV:"SSHPORT"`
	TargetHost      string `ENV:"TARGETHOST"`
	TargetPort      int    `ENV:"TARGETPORT"`
	SourceHost      string `ENV:"SOURCEHOST"`
	SourcePort      int    `ENV:"SOURCEPORT"`
}

func main() {

	// parse environment
	cfg := TunnelEnvironmentConfig{}
	env.Parse(cfg)

	authMethod, err := PublicKeyAuthMethod(cfg.CertificatePath, cfg.CerfiticatePass)
	if err != nil {
		log.Fatalf("Error reading/decrypting private key: %s", err)
		return
	}

	tunnel := &tunnel.Tunnel{
		Network: "upd",
		Config:  &ssh.ClientConfig{User: cfg.SSHUser, Auth: []ssh.AuthMethod{authMethod}},
		Proxy:   &tunnel.Endpoint{Host: cfg.SSHHost, Port: cfg.SSHPort},
		Source:  &tunnel.Endpoint{Host: cfg.SourceHost, Port: cfg.SourcePort},
		Target:  &tunnel.Endpoint{Host: cfg.TargetHost, Port: cfg.TargetPort},
	}

	if err := tunnel.Start(); err != nil {
		log.Fatalf("Error creating SSH Tunnel: %s", err)
	}
}
