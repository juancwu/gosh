package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func parseDestination(args []string) (user, host string) {
	argsWithParams := map[string]bool{
		"-b": true, "-c": true, "-D": true, "-E": true, "-e": true,
		"-F": true, "-I": true, "-i": true, "-L": true, "-l": true,
		"-m": true, "-O": true, "-o": true, "-p": true, "-R": true,
		"-S": true, "-W": true, "-w": true,
	}

	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "-") {
			flag := arg
			if len(arg) > 2 && !strings.Contains(arg, "=") {
				flag = "-" + string(arg[len(arg)-1])
			}

			if argsWithParams[flag] {
				skipNext = true
			}
			continue
		}

		parts := strings.SplitN(arg, "@", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return "", parts[0] // No user specified, just host
	}
	return "", ""
}

func checkAndEncryptKey(pemData []byte) ([]byte, error) {
	_, err := ssh.ParseRawPrivateKey(pemData)
	if err != nil {
		// Key is encrypted, so there is no more further action needed
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			return pemData, nil
		}
		return nil, err
	}

	fmt.Println("Warning: This key is unencrypted.")
	fmt.Print("Would you like to encrypt it before storing? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return pemData, nil
	}

	fmt.Print("Enter new passphrase: ")
	bytePass, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read passphrase: %w", err)
	}
	passphrase := string(bytePass)

	fmt.Print("Confirm passphrase: ")
	bytePassConfirm, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("fialed to read passphrase: %w", err)
	}
	fmt.Println()

	if passphrase != string(bytePassConfirm) {
		return nil, fmt.Errorf("passphrases do not match")
	}

	tempDir, err := os.MkdirTemp("", "gosh-encrypt")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	tempKeyPath := filepath.Join(tempDir, "gosh_temp_key")
	if err := os.WriteFile(tempKeyPath, pemData, 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command("ssh-keygen", "-p", "-f", tempKeyPath, "-P", "", "-N", passphrase, "-Z", "aes256-ctr")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ssh-keygen failed: %s: %s", err, string(output))
	}

	encryptedData, err := os.ReadFile(tempKeyPath)
	if err != nil {
		return nil, err
	}

	fmt.Println("Key encrypted successfully (AES-256-CTR).")

	return encryptedData, nil
}
