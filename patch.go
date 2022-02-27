package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type VersionPrevPatchProvider struct {
	baseFile *PatchFile
}

func NewVersionPatchProvider(baseFile *PatchFile) (*VersionPrevPatchProvider, error) {
	return &VersionPrevPatchProvider{
		baseFile: baseFile,
	}, nil
}

func (pp *VersionPrevPatchProvider) ID() uuid.UUID {
	return uuid.Nil
}

func (pp *VersionPrevPatchProvider) Changed() []BaseEntry {
	return pp.baseFile.Changed
}

func patch(cfg *Config, tagName string, srcDir string) error {
	baseFile, err := fetchBase(tagName, cfg.Backend)
	if err != nil {
		return err
	}

	pp, err := NewVersionPatchProvider(baseFile)
	if err != nil {
		return err
	}

	return createStagedVersionOrPatch(cfg, srcDir, pp)
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
