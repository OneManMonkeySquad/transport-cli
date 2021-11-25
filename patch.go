package main

import (
	"bytes"
	"compress/zlib"
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

func patch(cfg *Config) {
	if len(os.Args) <= 3 {
		fmt.Println("tp patch {tag} {directory}")
		return
	}

	tagName := os.Args[2]
	srcDir := os.Args[3]

	baseFile, err := fetchBase(tagName, cfg.Backend)
	if err != nil {
		log.Fatal(err)
		return
	}

	os.RemoveAll("staging")
	os.Mkdir("staging", 0777)

	patch, err := createPatch(cfg, srcDir, baseFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	writeToJsonFile(patch, "staging/staged.json")
}

func createPatch(cfg *Config, srcDir string, baseFile *PatchFile) (*PatchFile, error) {
	patch := PatchFile{
		Version: 1,
		ID:      uuid.New(),
		BaseID:  baseFile.ID,
	}

	existingFileSet := make(map[string]struct{})

	for _, baseEntry := range baseFile.Changed {
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

		existingFileSet[baseEntry.FileName] = struct{}{}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		if hashStr != baseEntry.Hash {
			changed, err := add(cfg, hashStr, baseEntry.FileName, content)
			if err != nil {
				return nil, err
			}

			patch.Changed = append(patch.Changed, *changed)
		}
	}

	err := foo(cfg, srcDir, existingFileSet, &patch)
	if err != nil {
		return nil, err
	}

	if len(patch.Changed) == 0 && len(patch.Deleted) == 0 {
		return nil, errors.New("no changes")
	}

	return &patch, nil
}

func foo(cfg *Config, srcDir string, existingFileSet map[string]struct{}, patch *PatchFile) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, exists := existingFileSet[file.Name()]
		if exists {
			continue
		}

		if file.IsDir() {
			err := foo(cfg, filepath.Join(srcDir, file.Name()), existingFileSet, patch)
			if err != nil {
				return err
			}

			continue
		}

		filePath := filepath.Join(srcDir, file.Name())

		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		changed, err := add(cfg, hashStr, file.Name(), content)
		if err != nil {
			return err
		}

		patch.Changed = append(patch.Changed, *changed)
	}

	return nil
}

func add(cfg *Config, hashStr string, fileName string, content []byte) (*BaseEntry, error) {
	buf := new(bytes.Buffer)
	{
		zlibWriter := zlib.NewWriter(buf)
		defer zlibWriter.Close()

		_, err := zlibWriter.Write(content)
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

		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		f.Write(buf.Next(cfg.ChunkSize()))
	}

	var changed BaseEntry
	changed.FileName = fileName
	changed.Hash = hashStr
	changed.AdditionalChunks = numChunks - 1

	return &changed, nil
}

func fetchBase(tagName string, backend Backend) (*PatchFile, error) {
	db, err := downloadDatabase(backend)
	if err == os.ErrNotExist {
		return nil, errors.New("no database found; make sure you have uploaded at least one base patch")
	} else if err != nil {
		return nil, err
	}

	tag := db.findTag(tagName)
	if tag == uuid.Nil {
		return nil, fmt.Errorf("tag '%v' not found", tagName)
	}

	entry := db.findEntry(tag)
	if entry == nil {
		return nil, fmt.Errorf("entry for tag '%v' not found; existing database not consistent", tagName)
	}

	entryContent, err := backend.DownloadFile(entry.ID.String() + ".json")
	if err != nil {
		return nil, err
	}

	var baseFile PatchFile
	if err = json.Unmarshal(entryContent, &baseFile); err != nil {
		return nil, err
	}

	if baseFile.Version != 1 {
		return nil, errors.New("base file has wrong version")
	}

	return &baseFile, nil
}
