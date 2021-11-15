package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/google/uuid"
)

type Backend interface {
	UploadFile(fileName string, data []byte) error
	DownloadFile(fileName string) ([]byte, error)
	Close()
}

// Entries
type BaseEntry struct {
	FileName string
	Hash     string
}

type DeletedEntry struct {
	FileName string
}

// Files
type BaseFile struct {
	Version int
	ID      uuid.UUID
	Entries []BaseEntry
}

type PatchFile struct {
	Version int
	ID      uuid.UUID
	BaseID  uuid.UUID
	// Contains new and changed files
	Changed []BaseEntry
	Deleted []DeletedEntry
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func writeToJsonFile(v interface{}, path string) {
	str, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile(path, str, 0644)
}

func readBaseFile(path string) BaseFile {
	fileData, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var baseFile BaseFile
	err = json.Unmarshal(fileData, &baseFile)
	if err != nil {
		log.Fatal(err)
	}

	if baseFile.Version != 1 {
		log.Fatal("Base file has wrong version")
	}
	return baseFile
}

func readPatchFile(path string) PatchFile {
	var patchFile PatchFile
	{
		fileData, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(fileData, &patchFile)
		if err != nil {
			log.Fatal(err)
		}

		if patchFile.Version != 1 {
			log.Fatal("Patch file has wrong version")
		}
	}
	return patchFile
}
