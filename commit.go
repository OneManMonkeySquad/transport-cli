package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

func commit(backend Backend, tagName string) error {
	if strings.ContainsAny(tagName, " .:;'#+*~") {
		return fmt.Errorf("invalid tag name '%v'", tagName)
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
		return err
	}

	// Make sure entry is unique
	for _, entry := range db.Entries {
		if entry.ID == newEntryID {
			return errors.New("entry already exists")
		}
	}

	// Update/Insert tag
	{
		foundTag := false
		for i, tag := range db.Tags {
			if tag.Name == tagName {
				db.Tags[i].ID = newEntryID
				foundTag = true
				break
			}
		}
		if !foundTag {
			newTag := Tag{
				Name: tagName,
				ID:   newEntryID,
			}
			db.Tags = append(db.Tags, newTag)

			fmt.Printf("Added new tag '%v'\n", tagName)
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
			return err
		}

		if err = backend.UploadFile(dataFile, data); err != nil {
			return err
		}
	}

	// Upload patch
	{
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		err = backend.UploadFile(newEntryID.String()+".json", data)
		if err != nil {
			return err
		}
	}

	// Upload DB
	if err = uploadDatabase(db, backend); err != nil {
		return err
	}

	// Remove patch
	os.Remove(filePath)

	for _, dataFile := range dataFiles {
		os.Remove("staging/" + dataFile)
	}

	return nil
}
