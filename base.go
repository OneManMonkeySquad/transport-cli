package main

import (
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func base() {
	if len(os.Args) <= 2 {
		fmt.Println("tp base {directory}")
		return
	}

	srcDir := os.Args[2]

	var baseFile BaseFile
	baseFile.Version = 1
	baseFile.ID = uuid.New()

	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, file := range files {
		entry, err := processBaseFile(srcDir, file.Name())
		if err != nil {
			log.Fatal(err)
			return
		}

		baseFile.Entries = append(baseFile.Entries, *entry)
	}

	writeToJsonFile(baseFile, "staging/"+baseFile.ID.String()+".json")

	fmt.Println("base:" + baseFile.ID.String())
}

func processBaseFile(srcDir string, fileName string) (*BaseEntry, error) {
	filePath := filepath.Join(srcDir, fileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// Compress and write data blob
	{
		f, err := os.Create("staging/" + hashStr)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		w := zlib.NewWriter(f)
		defer w.Close()

		_, err = w.Write(content)
		if err != nil {
			return nil, err
		}
	}

	// Add entry
	entry := BaseEntry{
		FileName: fileName,
		Hash:     hashStr,
	}
	return &entry, nil
}
