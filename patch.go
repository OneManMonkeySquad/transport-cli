package main

import (
	"encoding/json"
	"errors"
	"fmt"

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
	return pp.baseFile.ID
}

func (pp *VersionPrevPatchProvider) Changed() []BaseEntry {
	return pp.baseFile.Changed
}

func patch(cfg *Config, tagName string, srcDir string) error {
	baseFile, err := fetchBase(tagName, cfg.dataHive, cfg.metaHive)
	if err != nil {
		return err
	}

	pp, err := NewVersionPatchProvider(baseFile)
	if err != nil {
		return err
	}

	return createStagedVersionOrPatch(cfg, srcDir, pp)
}

func fetchBase(tagName string, dataHive DataHive, metaHive MetaHive) (*PatchFile, error) {
	tag, err := metaHive.FindTagByName(tagName)
	if err != nil {
		return nil, fmt.Errorf("tag '%v' not found", tagName)
	}

	entryContent, err := dataHive.DownloadFile(tag.Id.String() + ".json")
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
