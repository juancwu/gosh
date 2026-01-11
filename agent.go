package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func startEphemeralAgent(pemData []byte, target string) (string, func(), error) {
	var key interface{}
	var err error

	key, err = ssh.ParseRawPrivateKey(pemData)

	if err != nil {
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			fmt.Printf("\033[1;32m? Gosh:\033[0m Key for \033[1m%s\033[0m is encrypted. Enter passphrase: ", target)
			pass, readErr := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if readErr != nil {
				return "", nil, fmt.Errorf("failed to read password: %w", err)
			}

			key, err = ssh.ParseRawPrivateKeyWithPassphrase(pemData, pass)
			if err != nil {
				return "", nil, fmt.Errorf("incorrect passphrase or invalid key: %w", err)
			}
		} else {
			return "", nil, fmt.Errorf("invalid key format: %w", err)
		}
	} else {
		fmt.Printf("Using unencrypted key for %s\n", target)
	}

	keyring := agent.NewKeyring()
	addedKey := agent.AddedKey{
		PrivateKey:   key,
		Comment:      "gosh-ephemeral",
		LifetimeSecs: 60,
	}
	if err := keyring.Add(addedKey); err != nil {
		return "", nil, fmt.Errorf("failed to add key to agent: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "gosh-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to make temporary directory: %w", err)
	}
	sockPath := filepath.Join(tempDir, "agent.sock")

	l, err := net.Listen("unix", sockPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen on socket: %w", err)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go agent.ServeAgent(keyring, conn)
		}
	}()

	cleanup := func() {
		l.Close()
		os.RemoveAll(tempDir)
	}
	return sockPath, cleanup, nil
}
