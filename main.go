package main

import (
	"flag"
	"fmt"
	"os"
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
}
