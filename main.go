package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [gosh-flags] [user@host] [ssh-flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nManagement Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] list-keys\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] add-key <user_pattern> <host_pattern> <path_to_key>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] update-key <user_pattern> <host_pattern> <path_to_key>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] delete-key <id>\n", os.Args[0])
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

	switch cmd {
	case "add-key":
		if argc != 4 {
			fmt.Println("Error: Incorrect arguments for add-key.")
			fmt.Println("Try: gosh [--db path] add-key <user_pattern> <host_pattern> <path_to_key>")
			os.Exit(1)
		}

		err := handleAddKey(*dbPath, args[1], args[2], args[3])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)

	case "list-keys":
		err := handleListKey(*dbPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)

	case "update-key":
		if argc != 4 {
			fmt.Println("Usage: gosh update-key <user_pattern> <host_pattern> <path_to_key>")
			os.Exit(1)
		}

		err := handleUpdateKey(*dbPath, args[1], args[2], args[3])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)

	case "delete-key":
		if argc != 2 {
			fmt.Println("Usage: gosh delete-key <id>")
			os.Exit(1)
		}

		id, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		err = handleDeleteKey(*dbPath, int(id))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		startSSH(*dbPath, args)
	}
}
