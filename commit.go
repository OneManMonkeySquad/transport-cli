package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

func commit(backend Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp commit {tag} {guid}")
		return
	}

	updateTagName := os.Args[2]
	if strings.ContainsAny(updateTagName, " .:;'#+*~") {
		fmt.Println("Tag name contains invalid chars", updateTagName)
		return
	}

	fileTypeAndGuid := os.Args[3]

	var newEntryID uuid.UUID
	var newBaseID uuid.UUID
	var filePath string
	var dataFiles []string
	{
		if strings.HasPrefix(fileTypeAndGuid, "base:") {
			filePath = "staging/" + fileTypeAndGuid[5:] + ".json"
			base := readBaseFile(filePath)
			newEntryID = base.ID
			newBaseID = uuid.Nil
			for _, entry := range base.Entries {
				dataFiles = append(dataFiles, hex.EncodeToString(entry.SHA256Hash[:]))
			}
		} else if strings.HasPrefix(fileTypeAndGuid, "patch:") {
			filePath = "staging/" + fileTypeAndGuid[6:] + ".json"
			patch := readPatchFile(filePath)
			newEntryID = patch.ID
			newBaseID = patch.BaseID
			for _, entry := range patch.Changed {
				dataFiles = append(dataFiles, hex.EncodeToString(entry.SHA256Hash[:]))
			}
		} else {
			fmt.Println("Not a valid patch identifier", fileTypeAndGuid)
			return
		}
	}

	//
	db, err := downloadDatabase(backend)
	if _, ok := err.(*os.PathError); ok {
		fmt.Println("Existing patch database not found, creating one...")
		db = &Database{}
	} else if err == os.ErrNotExist {
		fmt.Println("Existing patch database not found, creating one...")
		db = &Database{}
	} else if err != nil {
		log.Fatal(err)
		return
	}

	// Make sure entry is unique
	for _, entry := range db.Entries {
		if entry.ID == newEntryID {
			log.Fatal("Entry already exists")
			return
		}
	}

	// Update/Insert tag
	{
		foundTag := false
		for i, tag := range db.Tags {
			if tag.Name == updateTagName {
				db.Tags[i].ID = newEntryID
				foundTag = true
				break
			}
		}
		if !foundTag {
			newTag := Tag{
				Name: updateTagName,
				ID:   newEntryID,
			}
			db.Tags = append(db.Tags, newTag)

			fmt.Println("Added new tag", updateTagName)
		}
	}

	// Insert entry
	newEntry := DatabaseEntry{
		ID:     newEntryID,
		BaseID: newBaseID,
	}
	db.Entries = append(db.Entries, newEntry)

	// Upload datas
	for _, dataFile := range dataFiles {
		filePath := "staging/" + dataFile

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = backend.UploadFile(dataFile, data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Upload patch
	{
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = backend.UploadFile(newEntryID.String()+".json", data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Upload DB
	err = uploadDatabase(db, backend)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Remove patch
	os.Remove(filePath)

	for _, dataFile := range dataFiles {
		filePath := "staging/" + dataFile
		os.Remove(filePath)
	}
}
