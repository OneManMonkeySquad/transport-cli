package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

func commit(cfg *Config, tagName string) error {
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

	// Make sure entry is unique
	{
		entry, err := cfg.metaHive.FindEntry(newEntryID)
		if err != nil {
			return err
		}
		if entry != uuid.Nil {
			return errors.New("entry exists already")
		}
	}

	// Upload datas
	for _, dataFile := range dataFiles {
		data, err := os.ReadFile("staging/" + dataFile)
		if err != nil {
			return err
		}

		fmt.Println("Uploading", dataFile, "...")
		if err = cfg.dataHive.UploadFile(dataFile, data); err != nil {
			return err
		}
	}

	// Upload patch
	{
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		err = cfg.dataHive.UploadFile(newEntryID.String()+".json", data)
		if err != nil {
			return err
		}
	}

	// Update/Insert tag
	cfg.metaHive.UpdateTag(tagName, newEntryID)

	// Insert entry
	cfg.metaHive.AddEntry(newEntryID, newBaseID)

	// Remove patch
	os.Remove(filePath)

	for _, dataFile := range dataFiles {
		os.Remove("staging/" + dataFile)
	}

	return nil
}
