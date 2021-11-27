package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FlatPatch struct {
	Entries []BaseEntry
	Deleted []DeletedEntry
}

func restore(backend Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp restore {tag} {dir}")
		return
	}

	db, err := downloadDatabase(backend)
	if err != nil {
		fmt.Println("Patch database not found.")
		return
	}

	tagName := os.Args[2]
	path := os.Args[3]
	fmt.Printf("Restoring '%s'...\n", tagName)

	head := db.findTag(tagName)

	restoreChain := findRestoreChain(db, head)

	// Now, instead of just going through patches, we collapse them into one.
	// This way we don't write a single file multiple times or write and then delete a file.
	flatPatch, err := flattenRestoreChain(db, restoreChain, backend)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, entry := range flatPatch.Entries {
		filePath := filepath.Join(path, entry.FileName)

		hashStr := ""

		existingContent, err := os.ReadFile(filePath)
		if err == nil {
			hash := sha256.Sum256(existingContent)
			hashStr = hex.EncodeToString(hash[:])
		}

		if hashStr != entry.Hash {
			var buffer = new(bytes.Buffer)
			for i := 0; i < entry.AdditionalChunks+1; i += 1 {
				name := entry.Hash
				if i > 0 {
					name = fmt.Sprintf("%s_%d", name, i)
				}

				newContent, err := backend.DownloadFile(name)
				if err != nil {
					log.Fatal(err)
					return
				}

				buffer.Write(newContent)
			}

			// Create directory
			dirName := filepath.Dir(filePath)
			if _, err := os.Stat(dirName); err != nil {
				if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
					log.Fatal(err)
					return
				}
			}

			// Write file
			file, err := os.Create(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			zlibReader, err := zlib.NewReader(buffer)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer zlibReader.Close()

			io.Copy(file, zlibReader)
		}
	}
	for _, entry := range flatPatch.Deleted {
		filePath := filepath.Join(path, entry.FileName)
		os.Remove(filePath)
	}
}

func findRestoreChain(db *Database, head uuid.UUID) []DatabaseEntry {
	var restoreChain []DatabaseEntry
	{
		var currentHead = head
		for currentHead != uuid.Nil {
			entry := db.findEntry(currentHead)
			if entry == nil {
				log.Fatal("Patch not found")
			}

			restoreChain = append(restoreChain, *entry)
			currentHead = entry.BaseID
		}
	}

	for i, j := 0, len(restoreChain)-1; i < j; i, j = i+1, j-1 {
		restoreChain[i], restoreChain[j] = restoreChain[j], restoreChain[i]
	}

	return restoreChain
}

func flattenRestoreChain(db *Database, restoreChain []DatabaseEntry, persistence Backend) (*FlatPatch, error) {
	entryMap := make(map[string]BaseEntry)
	deletedMap := make(map[string]DeletedEntry)

	for i, entry := range restoreChain {
		patchContent, err := persistence.DownloadFile(entry.ID.String() + ".json")
		if err != nil {
			return nil, err
		}

		if i == 0 {
			var baseFile PatchFile
			err = json.Unmarshal(patchContent, &baseFile)
			if err != nil {
				return nil, err
			}

			for _, entry := range baseFile.Changed {
				entryMap[entry.FileName] = entry
			}
		} else {
			var patchFile PatchFile
			err = json.Unmarshal(patchContent, &patchFile)
			if err != nil {
				return nil, err
			}

			for _, entry := range patchFile.Changed {
				entryMap[entry.FileName] = entry
				delete(deletedMap, entry.FileName)
			}
			for _, entry := range patchFile.Deleted {
				delete(entryMap, entry.FileName)
				deletedMap[entry.FileName] = entry
			}
		}
	}

	var result FlatPatch

	result.Entries = make([]BaseEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		result.Entries = append(result.Entries, entry)
	}

	result.Deleted = make([]DeletedEntry, 0, len(deletedMap))
	for _, entry := range deletedMap {
		result.Deleted = append(result.Deleted, entry)
	}

	return &result, nil
}
