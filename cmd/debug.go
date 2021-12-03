package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/net/proxy"
)

func main() {
	var (
		ctx     = context.Background()
		user    = "git"
		addr    = "github.com:22"
		network = "tcp"
		err     error
	)

	var hostKeyCallback ssh.HostKeyCallback
	{
		homeDirPath, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get UserHomeDir: %v", err)
		}

		knownHostFile := filepath.Join(homeDirPath, "/.ssh/known_hosts")
		log.Printf("knownHostFile: %s\n", knownHostFile)

		hostKeyCallback, err = knownhosts.New(knownHostFile)
		if err != nil {
			log.Fatalf("Failed to create host key callback: %v", err)
		}
	}

	var auth []ssh.AuthMethod
	{
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			log.Fatalf("Failed to open SSH_AUTH_SOCK: %v", err)
		}

		agentClient := agent.NewClient(conn)
		auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}
	}

	config := &ssh.ClientConfig{
		Auth:            auth,
		User:            user,
		HostKeyCallback: hostKeyCallback,
	}

	conn, err := proxy.Dial(ctx, network, addr)
	if err != nil {
		log.Fatalf("Failed to dial %s %s\n", network, addr)
	}

	_, _, _, err = ssh.NewClientConn(conn, addr, config)
	if err != nil {
		log.Fatalf("Failed to establish ssh connection: %v", err)
	}

	log.Println("success")
}
