package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func base(cfg *Config) {
	if len(os.Args) <= 2 {
		fmt.Println("tp base {directory}")
		return
	}

	srcDir := os.Args[2]

	os.RemoveAll("staging")
	os.Mkdir("staging", 0777)

	var baseFile PatchFile
	baseFile.Version = 1
	baseFile.ID = uuid.New()

	err := processDir(cfg, srcDir, ".", &baseFile.Changed)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(baseFile.Changed) == 0 {
		log.Fatal("No entries - folder empty?")
		return
	}

	err = writeToJsonFile(baseFile, "staging/staged.json")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func processDir(cfg *Config, dir string, subDir string, entries *[]BaseEntry) error {
	fullDir := filepath.Join(dir, subDir)

	dirEntries, err := os.ReadDir(fullDir)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			entry, err := processBaseFile(cfg, dir, subDir, dirEntry.Name())
			if err != nil {
				return err
			}

			*entries = append(*entries, *entry)
		} else {
			newSubDir := filepath.Join(subDir, dirEntry.Name())
			if err = processDir(cfg, dir, newSubDir, entries); err != nil {
				return err
			}
		}
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
	buf := new(bytes.Buffer)
	{
		zlibWriter := zlib.NewWriter(buf)
		defer zlibWriter.Close()

		_, err = zlibWriter.Write(content)
		if err != nil {
			return nil, err
		}
	}

	numChunks := (buf.Len() / cfg.ChunkSize()) + 1

	for i := 0; i < numChunks; i += 1 {
		name := "staging/" + hashStr
		if i > 0 {
			name = fmt.Sprintf("%s_%d", name, i)
		}

		file, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = file.Write(buf.Next(cfg.ChunkSize()))
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
