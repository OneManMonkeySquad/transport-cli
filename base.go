package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func base(cfg *Config, srcDir string) error {
	os.RemoveAll("staging")
	os.Mkdir("staging", 0777)

	var baseFile PatchFile
	baseFile.Version = 1
	baseFile.ID = uuid.New()

	err := processDir(cfg, srcDir, ".", &baseFile.Changed)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if len(baseFile.Changed) == 0 {
		return errors.New("no entries - folder empty?")
	}

	err = writeToJsonFile(baseFile, "staging/staged.json")
	if err != nil {
		return err
	}

	return nil
}

func processDir(cfg *Config, dir string, subDir string, entries *[]BaseEntry) error {
	fullDir := filepath.Join(dir, subDir)

	dirEntries, err := os.ReadDir(fullDir)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			newSubDir := filepath.Join(subDir, dirEntry.Name())
			if err = processDir(cfg, dir, newSubDir, entries); err != nil {
				return err
			}
			continue
		}

		entry, err := processBaseFile(cfg, dir, subDir, dirEntry.Name())
		if err != nil {
			return err
		}

		*entries = append(*entries, *entry)
	}

	return nil
}

func processBaseFile(cfg *Config, dir string, subDir string, fileName string) (*BaseEntry, error) {
	filePath := filepath.Join(dir, subDir, fileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// Compress and write data blob
	compressedContent := new(bytes.Buffer)
	{
		zlibWriter := zlib.NewWriter(compressedContent)
		defer zlibWriter.Close()

		_, err = zlibWriter.Write(content)
		if err != nil {
			return nil, err
		}

		err = zlibWriter.Flush()
		if err != nil {
			return nil, err
		}
	}

	numChunks := (compressedContent.Len() / cfg.ChunkSize()) + 1
	if numChunks > 1024 {
		return nil, errors.New("too many chunks")
	}

	for i := 0; i < numChunks; i++ {
		name := "staging/" + hashStr
		if i > 0 {
			name = fmt.Sprintf("%s_%d", name, i)
		}

		chunk := compressedContent.Next(cfg.ChunkSize())
		err := os.WriteFile(name, chunk, 0666)
		if err != nil {
			return nil, err
		}
	}

	// Add entry
	entry := BaseEntry{
		FileName:         filepath.Join(subDir, fileName),
		Hash:             hashStr,
		AdditionalChunks: numChunks - 1,
	}
	return &entry, nil
}
