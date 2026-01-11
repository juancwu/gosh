package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func handleAddKey(dbPath, userPattern, hostPattern, keyPath string) error {
	db, err := initDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	pemData, err := readKeyFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}

	finalPemData, err := checkAndEncryptKey(pemData)
	if err != nil {
		return fmt.Errorf("failed to check and encrypt key: %w", err)
	}

	source := "Stdin"
	if keyPath != "-" {
		source = keyPath
	}

	err = addKey(db, finalPemData, userPattern, hostPattern, source)
	if err != nil {
		return err
	}

	fmt.Printf("Key for %s@%s imported successfully.\n", userPattern, hostPattern)

	return nil
}

func handleListKey(dbPath string) error {
	db, err := initDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	keys, err := listkeys(db)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "UD\tUser Pattern\tHost Pattern\tComment")
	fmt.Fprintln(w, "--\t------------\t------------\t-------")
	for _, k := range keys {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", k.ID, k.UserPattern, k.HostPattern, k.Comment)
	}
	w.Flush()
	return nil
}

func handleUpdateKey(dbPath, userPattern, hostPattern, keyPath string) error {
	db, err := initDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	err = updateKey(db, userPattern, hostPattern, keyPath)
	if err != nil {
		return err
	}

	return nil
}

func handleDeleteKey(dbPath string, id int) error {
	db, err := initDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	err = deleteKey(db, id)
	if err != nil {
		return err
	}

	return nil
}
