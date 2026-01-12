package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

type KeyRecord struct {
	ID           int    `json:"id"`
	HostPattern  string `json:"host_pattern"`
	UserPattern  string `json:"user_pattern"`
	EncryptedPem []byte `json:"encrypted_pem"` // Storea as base64, default marshal encoding
	Comment      string `json:"comment"`
}

type Storage struct {
	Keys []KeyRecord `json:"keys"`
}

func getStorePath(customPath string) string {
	if customPath != "" {
		return customPath
	}
	home, _ := os.UserHomeDir()
	localDataDir := filepath.Join(home, ".local", "share", "gosh")
	err := os.MkdirAll(localDataDir, 0700)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Failed to create local data direction '", localDataDir, "': ", err)
		fmt.Fprintln(os.Stderr, "Warning: Putting database in current working directory.")
		return "./keys.json"
	}
	return filepath.Join(localDataDir, "keys.json")
}

func loadStore(storePath string) (*Storage, error) {
	data, err := os.ReadFile(storePath)
	if os.IsNotExist(err) {
		return &Storage{Keys: []KeyRecord{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var store Storage
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("corrupt store file: %w", err)
	}
	return &store, nil
}

func saveStore(storePath string, store *Storage) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storePath, data, 0600)
}

func findKey(storePath, user, host string) ([]byte, error) {
	store, err := loadStore(getStorePath(storePath))
	if err != nil {
		return nil, err
	}

	sort.Slice(store.Keys, func(i, j int) bool {
		return len(store.Keys[i].HostPattern) > len(store.Keys[j].HostPattern)
	})

	for _, k := range store.Keys {
		hostMatched, _ := filepath.Match(k.HostPattern, host)
		userMatched, _ := filepath.Match(k.UserPattern, user)
		if hostMatched && userMatched {
			return k.EncryptedPem, nil
		}
	}

	return nil, fmt.Errorf("no matching key found")
}

func addKey(storePath, userPattern, hostPattern, keyPath string) error {
	realPath := getStorePath(storePath)
	store, err := loadStore(realPath)
	if err != nil {
		return err
	}

	pemData, err := readKeyFile(keyPath)
	if err != nil {
		return err
	}

	finalPemData, err := checkAndEncryptKey(pemData)
	if err != nil {
		return err
	}

	source := "Stdin"
	if keyPath != "-" {
		source = keyPath
	}

	maxID := 0
	for _, k := range store.Keys {
		if k.ID > maxID {
			maxID = k.ID
		}
	}

	newKey := KeyRecord{
		ID:           maxID + 1,
		HostPattern:  hostPattern,
		UserPattern:  userPattern,
		EncryptedPem: finalPemData,
		Comment:      "Imported from " + source,
	}

	store.Keys = append(store.Keys, newKey)

	if err := saveStore(realPath, store); err != nil {
		return err
	}

	fmt.Printf("Key for %s@%s imported successfully (ID: %d).\n", userPattern, hostPattern, newKey.ID)

	return nil
}

func listkeys(storePath string) ([]KeyRecord, error) {
	store, err := loadStore(getStorePath(storePath))
	if err != nil {
		return nil, err
	}

	return store.Keys, nil
}

func updateKey(storePath string, keyID int, userPattern, hostPattern, keyPath string) error {
	realPath := getStorePath(storePath)
	store, err := loadStore(realPath)
	if err != nil {
		return err
	}

	pemData, err := readKeyFile(keyPath)
	if err != nil {
		return err
	}

	finalPemData, err := checkAndEncryptKey(pemData)
	if err != nil {
		return err
	}

	source := "Stdin"
	if keyPath != "-" {
		source = keyPath
	}

	found := false
	for i, k := range store.Keys {
		if k.ID == keyID {
			store.Keys[i].HostPattern = hostPattern
			store.Keys[i].UserPattern = userPattern
			store.Keys[i].EncryptedPem = finalPemData
			store.Keys[i].Comment = "Updated from " + source
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("no key found with ID %d.\n", keyID)
		return nil
	}

	err = saveStore(realPath, store)
	if err != nil {
		return fmt.Errorf("failed to save store: %w", err)
	}

	fmt.Printf("key %d updated successfully.\n", keyID)

	return nil
}

func deleteKey(storePath string, id int) error {
	realPath := getStorePath(storePath)
	store, err := loadStore(realPath)
	if err != nil {
		return err
	}

	newKeys := []KeyRecord{}
	found := false
	for _, k := range store.Keys {
		if k.ID == id {
			found = true
			continue
		}
		newKeys = append(newKeys, k)
	}

	if !found {
		fmt.Printf("no key found with ID %d.\n", id)
		return nil
	}

	store.Keys = newKeys
	err = saveStore(realPath, store)
	if err != nil {
		return err
	}

	fmt.Printf("key %d deleted.\n", id)

	return nil
}
