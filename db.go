package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

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
