package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Tag struct {
	Name string
	ID   uuid.UUID
}

type DatabaseEntry struct {
	ID     uuid.UUID
	BaseID uuid.UUID
}

type Database struct {
	Tags    []Tag
	Entries []DatabaseEntry
}

func (db *Database) findTag(name string) uuid.UUID {
	for _, tag := range db.Tags {
		if tag.Name == name {
			return tag.ID
		}
	}
	return uuid.Nil
}

func (db *Database) findEntry(id uuid.UUID) *DatabaseEntry {
	for _, entry := range db.Entries {
		if entry.ID == id {
			return &entry
		}
	}
	return nil
}

func downloadDatabase(persistence Backend) (*Database, error) {
	dbContent, err := persistence.DownloadFile("db.json")
	if err != nil {
		return nil, err
	}

	var db Database
	err = json.Unmarshal(dbContent, &db)
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func uploadDatabase(db *Database, persistence Backend) error {
	content, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return persistence.UploadFile("db.json", content)
}
