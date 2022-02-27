package main

import "github.com/google/uuid"

type NullPrevPatchProvider struct {
}

func (pp *NullPrevPatchProvider) ID() uuid.UUID {
	return uuid.Nil
}

func (pp *NullPrevPatchProvider) Changed() []BaseEntry {
	return []BaseEntry{}
}

func base(cfg *Config, srcDir string) error {
	pp := &NullPrevPatchProvider{}
	return createStagedVersionOrPatch(cfg, srcDir, pp)
}
