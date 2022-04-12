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

type BaseEntry struct {
	FileName         string
	Hash             string
	AdditionalChunks int `json:"AdditionalChunks,omitempty"`
}

type PatchFile struct {
	Version int
	ID      uuid.UUID
	BaseID  uuid.UUID
	// Contains new and changed files
	Changed []BaseEntry
	Deleted []DeletedEntry
}

type DeletedEntry struct {
	FileName string
}

type PrevPatchProvider interface {
	ID() uuid.UUID
	Changed() []BaseEntry
}

func createStagedVersionOrPatch(cfg *Config, srcDir string, pp PrevPatchProvider) error {
	os.RemoveAll(".staging")
	os.Mkdir(".staging", 0777)

	patch, err := createPatch(cfg, srcDir, pp)
	if err != nil {
		return err
	}

	return writeToJsonFile(patch, ".staging/staged.json")
}

func createPatch(cfg *Config, srcDir string, pp PrevPatchProvider) (*PatchFile, error) {
	patch := PatchFile{
		Version: 1,
		ID:      uuid.New(),
		BaseID:  pp.ID(),
	}

	existingFileSet := make(map[string]struct{})

	for _, baseEntry := range pp.Changed() {
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
			changed, err := processPatchFile(cfg, hashStr, baseEntry.FileName, content)
			if err != nil {
				return nil, err
			}

			patch.Changed = append(patch.Changed, *changed)
		}
	}

	err := processPatchDir(cfg, srcDir, "", existingFileSet, &patch)
	if err != nil {
		return nil, err
	}

	if len(patch.Changed) == 0 && len(patch.Deleted) == 0 {
		return nil, errors.New("no changes")
	}

	return &patch, nil
}

func processPatchDir(cfg *Config, srcDir string, currentSubDir string, existingFileSet map[string]struct{}, patch *PatchFile) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, exists := existingFileSet[filepath.Join(currentSubDir, file.Name())]
		if exists {
			continue
		}

		if file.IsDir() {
			err := processPatchDir(cfg, filepath.Join(srcDir, file.Name()), filepath.Join(currentSubDir, file.Name()), existingFileSet, patch)
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

		changed, err := processPatchFile(cfg, hashStr, filepath.Join(currentSubDir, file.Name()), content)
		if err != nil {
			return err
		}

		patch.Changed = append(patch.Changed, *changed)
	}

	return nil
}

func processPatchFile(cfg *Config, hashStr string, fileName string, content []byte) (*BaseEntry, error) {
	compressedContent := new(bytes.Buffer)
	{
		zlibWriter := zlib.NewWriter(compressedContent)
		defer zlibWriter.Close()

		_, err := zlibWriter.Write(content)
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
		name := ".staging/" + hashStr
		if i > 0 {
			name = fmt.Sprintf("%s_%d", name, i)
		}

		chunk := compressedContent.Next(cfg.ChunkSize())

		if err := os.WriteFile(name, chunk, 0666); err != nil {
			return nil, err
		}
	}

	changed := BaseEntry{
		FileName:         fileName,
		Hash:             hashStr,
		AdditionalChunks: numChunks - 1,
	}
	return &changed, nil
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func writeToJsonFile(v interface{}, path string) error {
	str, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, str, 0644)
}

func readPatchFile(path string) PatchFile {
	var patchFile PatchFile
	{
		fileData, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(fileData, &patchFile)
		if err != nil {
			log.Fatal(err)
		}

		if patchFile.Version != 1 {
			log.Fatal("Patch file has wrong version")
		}
	}
	return patchFile
}
