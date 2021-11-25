package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

func commit(backend Backend) {
	if len(os.Args) <= 2 {
		fmt.Println("tp commit {tag}")
		return
	}

	updateTagName := os.Args[2]
	if strings.ContainsAny(updateTagName, " .:;'#+*~") {
		fmt.Println("Tag name contains invalid chars", updateTagName)
		return
	}

	var newEntryID uuid.UUID
	var newBaseID uuid.UUID
	filePath := "staging/staged.json"
	var dataFiles []string
	{
		patch := readPatchFile(filePath)
		newEntryID = patch.ID
		newBaseID = patch.BaseID
		for _, entry := range patch.Changed {
			for i := 0; i < entry.AdditionalChunks+1; i += 1 {
				name := entry.Hash
				if i > 0 {
					name = fmt.Sprintf("%s_%d", name, i)
				}

				dataFiles = append(dataFiles, name)
			}
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
		data, err := os.ReadFile("staging/" + dataFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err = backend.UploadFile(dataFile, data); err != nil {
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
	if err = uploadDatabase(db, backend); err != nil {
		fmt.Println(err)
		return
	}

	// Remove patch
	os.Remove(filePath)

	for _, dataFile := range dataFiles {
		os.Remove("staging/" + dataFile)
	}
}
