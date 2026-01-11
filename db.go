package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type KeyRecord struct {
	ID          int
	HostPattern string
	UserPattern string
	Comment     string
}

func getDBPath() string {
	home, _ := os.UserHomeDir()
	localDataDir := filepath.Join(home, ".local", "share", "gosh")
	err := os.MkdirAll(localDataDir, 0700)
	if err != nil {
		fmt.Println("[WARNING] Failed to create local data direction '", localDataDir, "': ", err)
		fmt.Println("[WARNING] Putting database in current working directory.")
		return "./gosh.db"
	}
	return filepath.Join(localDataDir, "gosh.db")
}

func initDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = getDBPath()
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func findKey(db *sql.DB, user, host string) ([]byte, error) {
	query := `
	SELECT encrypted_pem FROM keys
	WHERE (host_pattern = ? OR host_pattern = '*') AND (user_pattern = ? OR user_pattern = '*')
	ORDER BY id DESC LIMIT 1;
	`

	var pemData []byte
	err := db.QueryRow(query, host, user).Scan(&pemData)
	if err != nil {
		return nil, err
	}

	return pemData, nil
}

func addKey(db *sql.DB, userPattern, hostPattern, keyPath string) error {
	pemData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	query := `
	INSERT INTO keys (host_pattern, user_pattern, encrypted_pem, comment)
	VALUES (?, ?, ? ,?);
	`
	_, err = db.Exec(query, hostPattern, userPattern, pemData, "Imported from "+keyPath)
	if err != nil {
		return fmt.Errorf("failed to add key: %w", err)
	}

	return nil
}

func listkeys(db *sql.DB) ([]KeyRecord, error) {
	rows, err := db.Query("SELECT id, host_pattern, user_pattern, comment FROM keys;")
	if err != nil {
		return nil, fmt.Errorf("failed to query keys: %w", err)
	}
	defer rows.Close()

	var keys []KeyRecord
	for rows.Next() {
		var k KeyRecord
		if err := rows.Scan(&k.ID, &k.HostPattern, &k.UserPattern, &k.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func updateKey(db *sql.DB, userPattern, hostPattern, keyPath string) error {
	pemData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}

	res, err := db.Exec(
		"UPDATE keys SET encrypted_pem=?, comment=? WHERE user_pattern = ? AND host_pattern = ?;",
		pemData, "Updated from "+keyPath, userPattern, hostPattern,
	)
	if err != nil {
		return fmt.Errorf("failed to update key: %w", err)
	}

	rows, err := res.RowsAffected()
	if err == nil {
		if rows == 0 {
			fmt.Printf("No key found with user '%s' and host '%s'.\n", userPattern, hostPattern)
		} else {
			fmt.Printf("Key for %s@%s updated successfully.\n", userPattern, hostPattern)
		}
	} else {
		fmt.Println("Warning: could not verify update result.", err)
	}

	return nil
}

func deleteKey(db *sql.DB, id int) error {
	res, err := db.Exec("DELETE FROM keys WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	rows, err := res.RowsAffected()
	if err == nil {
		if rows == 0 {
			fmt.Printf("No key found with ID %d.\n", id)
		} else {
			fmt.Printf("Key with ID %d deleted.\n", id)
		}
	} else {
		fmt.Println("Warning: could not confirm key deletion.", err)
	}

	return nil
}
