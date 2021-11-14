package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func patch(backend Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp patch {tag} {directory}")
		return
	}

	db, err := downloadDatabase(backend)
	if err == os.ErrNotExist {
		fmt.Println("No Database found. Make sure you have uploaded at least one base patch.")
		return
	} else if err != nil {
		log.Fatal(err)
		return
	}

	tagName := os.Args[2]
	tag := db.findTag(tagName)
	if tag == uuid.Nil {
		log.Fatal("Tag not found", tagName)
		return
	}

	entry := db.findEntry(tag)
	if entry == nil {
		log.Fatal("Entry for tag not found. Existing database not consistent", tagName)
		return
	}

	entryContent, err := backend.DownloadFile(entry.ID.String() + ".json")
	if err != nil {
		log.Fatal(err)
		return
	}

	var baseFile BaseFile
	err = json.Unmarshal(entryContent, &baseFile)
	if err != nil {
		log.Fatal(err)
	}

	if baseFile.Version != 1 {
		log.Fatal("Base file has wrong version")
		return
	}

	patch, err := createPatch(os.Args[3], baseFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	writeJson(patch, "staging/"+patch.ID.String()+".json")
	fmt.Println("patch:" + patch.ID.String())
}

func createPatch(srcDir string, baseFile BaseFile) (*PatchFile, error) {
	var patch PatchFile
	patch.Version = 1
	patch.ID = uuid.New()
	patch.BaseID = baseFile.ID

	for _, baseEntry := range baseFile.Entries {
		filePath := filepath.Join(srcDir, baseEntry.FileName)

		stillExits, err := exists(filePath)
		if err != nil {
			return nil, err
		}

		if !stillExits {
			var deleted DeletedEntry
			deleted.FileName = baseEntry.FileName
			patch.Deleted = append(patch.Deleted, deleted)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(content)
		if hash != baseEntry.SHA256Hash {
			err = os.WriteFile("staging/"+hex.EncodeToString(hash[:]), content, 0644)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}

			var changed BaseEntry
			changed.FileName = baseEntry.FileName
			changed.SHA256Hash = hash
			patch.Changed = append(patch.Changed, changed)
			continue
		}
	}

	if len(patch.Changed) == 0 && len(patch.Deleted) == 0 {
		return nil, errors.New("no changes")
	}

	return &patch, nil
}
