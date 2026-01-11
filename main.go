package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [gosh-flags] [user@host] [ssh-flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nManagement Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] add-key <user_pattern> <host_pattern> <path_to_key>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	dbPath := flag.String("db", "", "Path to the keys database (optional)")
	flag.Parse()

	args := flag.Args()
	argc := len(args)

	if argc == 0 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := args[0]

	if cmd == "add-key" {
		if argc != 4 {
			fmt.Println("Error: Incorrect arguments for add-key.")
			fmt.Println("Try: gosh [--db path] add-key <user_pattern> <host_pattern> <path_to_key>")
			os.Exit(1)
		}

		db, err := initDB(*dbPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		defer db.Close()

		userPattern := args[1]
		hostPattern := args[2]
		keyPath := args[3]
		err = addKey(db, userPattern, hostPattern, keyPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	user, host := parseDestination(args)
	env := os.Environ()

	if host != "" {
		db, err := initDB(*dbPath)
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
