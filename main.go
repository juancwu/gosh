package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [gosh-flags] [user@host] [ssh-flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nManagement Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] list-keys\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] add-key <user_pattern> <host_pattern> <path_to_key>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] update-key <id> <user_pattern> <host_pattern> <path_to_key>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] delete-key <id>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	storePath := flag.String("store", "", "Path to the keys store (optional)")
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
			fmt.Println("Try: gosh [-store path] add-key <user_pattern> <host_pattern> <path_to_key>")
			os.Exit(1)
		}

		user := args[1]
		host := args[2]
		keyPath := args[3]

		err := addKey(*storePath, user, host, keyPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)

	case "list-keys":
		keys, err := listkeys(*storePath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "UD\tUser Pattern\tHost Pattern\tComment")
		fmt.Fprintln(w, "--\t------------\t------------\t-------")
		for _, k := range keys {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", k.ID, k.UserPattern, k.HostPattern, k.Comment)
		}
		w.Flush()

		os.Exit(0)

	case "update-key":
		if argc != 5 {
			fmt.Println("Usage: gosh update-key <id> <user_pattern> <host_pattern> <path_to_key>")
			os.Exit(1)
		}

		idStr := args[1]
		user := args[2]
		host := args[3]
		keyPath := args[4]

		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		err = updateKey(*storePath, int(id), user, host, keyPath)
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

		err = deleteKey(*storePath, int(id))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		startSSH(*storePath, args)
	}
}
