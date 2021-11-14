package main

import (
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
	{
		baseFile.Version = 1
		baseFile.ID = uuid.New()

		files, err := os.ReadDir(srcDir)
		if err != nil {
			log.Fatal(err)
			return
		}

		for _, file := range files {
			filePath := filepath.Join(srcDir, file.Name())

			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			hash := sha256.Sum256(content)
			hashStr := hex.EncodeToString(hash[:])

			err = os.WriteFile("staging/"+hashStr, content, 0644)
			if err != nil {
				log.Fatal(err)
				return
			}

			entry := BaseEntry{
				FileName:   file.Name(),
				SHA256Hash: hashStr,
			}
			baseFile.Entries = append(baseFile.Entries, entry)
		}
	}

	writeJson(baseFile, "staging/"+baseFile.ID.String()+".json")
	fmt.Println("base:" + baseFile.ID.String())
}
