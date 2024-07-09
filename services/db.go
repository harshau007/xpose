package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Name         string `yaml:"name"`
	ContainerID  string `yaml:"container_id"`
	ExternalPort string `yaml:"external_port"`
	InternalPort string `yaml:"internal_port"`
	Type         string `yaml:"type"`
	Source       string `yaml:"source"`
	PublicURL    string `yaml:"public_url"`
	PID          string `yaml:"pid"`
}

type Database []Entry

type Transaction struct {
	db       *Database
	filename string
	tempFile string
}

func NewTransaction(filename string) (*Transaction, error) {
	if err := ensureFileExists(filename); err != nil {
		return nil, fmt.Errorf("error ensuring file exists: %v", err)
	}

	db, err := readDatabase(filename)
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	tempFile := filepath.Join(dir, fmt.Sprintf("%s.tmp", filepath.Base(filename)))
	return &Transaction{
		db:       db,
		filename: filename,
		tempFile: tempFile,
	}, nil
}

func ensureFileExists(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Printf("File %s does not exist. Creating it with an empty database.", filename)
		emptyDB := &Database{}
		return writeDatabase(filename, emptyDB)
	}
	return err
}

func (t *Transaction) commit() error {
	// Write to temp file
	if err := writeDatabase(t.tempFile, t.db); err != nil {
		return fmt.Errorf("failed to write to temp file: %v", err)
	}

	// Rename temp file to original file (atomic operation)
	if err := os.Rename(t.tempFile, t.filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

func (t *Transaction) rollback() {
	os.Remove(t.tempFile)
}

func readDatabase(filename string) (*Database, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var db Database
	err = yaml.Unmarshal(data, &db)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	return &db, nil
}

func writeDatabase(filename string, db *Database) error {
	data, err := yaml.Marshal(db)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func (t *Transaction) CreateEntry(entry Entry) {
	*t.db = append(*t.db, entry)
	log.Printf("Created new entry: %s", entry.Name)
}

func (t *Transaction) ReadEntry(name string) (*Entry, bool) {
	for _, entry := range *t.db {
		if entry.Name == name {
			log.Printf("Found entry: %s", name)
			return &entry, true
		}
	}
	log.Printf("Entry not found: %s", name)
	return nil, false
}

func (t *Transaction) UpdateEntry(id string, updatedEntry Entry) bool {
	for i, entry := range *t.db {
		if strings.Contains(entry.ContainerID, id) {
			(*t.db)[i] = updatedEntry
			log.Printf("Updated entry: %s", id)
			return true
		}
	}
	log.Printf("Failed to update entry: %s (not found)", id)
	return false
}

func (t *Transaction) DeleteEntry(id string) bool {
	for i, entry := range *t.db {
		if strings.Contains(entry.ContainerID, id) {
			*t.db = append((*t.db)[:i], (*t.db)[i+1:]...)
			log.Printf("Deleted entry: %s", id)
			return true
		}
	}
	log.Printf("Failed to delete entry: %s (not found)", id)
	return false
}

func (t *Transaction) ReadEntryById(id string) (*Entry, bool) {
	for _, entry := range *t.db {
		if strings.Contains(entry.ContainerID, id) {
			return &entry, true
		}
	}
	return nil, false
}
