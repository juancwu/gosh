package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func startEphemeralAgent(pemData []byte, target string) (string, func(), error) {
	var key any
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

func startSSH(dbPath string, args []string) {
	user, host := parseDestination(args)
	env := os.Environ()

	if host != "" {
		db, err := initDB(dbPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		pemData, err := findKey(db, user, host)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		db.Close()

		targetName := host
		if user != "" {
			targetName = user + "@" + host
		}

		socketPath, cleanup, err := startEphemeralAgent(pemData, targetName)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		defer cleanup()

		newEnv := []string{"SSH_AUTH_SOCK=" + socketPath}
		for _, e := range env {
			if !strings.HasPrefix(e, "SSH_AUTH_SOCK=") {
				newEnv = append(newEnv, e)
			}
		}
		env = newEnv
	}

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	sshCmd := exec.Command(sshPath, args...)
	sshCmd.Env = env
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			if sshCmd.Process != nil {
				sshCmd.Process.Signal(sig)
			}
		}
	}()

	err = sshCmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
