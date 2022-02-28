package main

import (
	"github.com/OneManMonkeySquad/transport-cli/meta_hives"
	"github.com/google/uuid"
)

type MetaHive interface {
	Tags() ([]meta_hives.Tag, error)
	FindTagByName(name string) (*meta_hives.Tag, error)
	UpdateTag(name string, newId uuid.UUID) error

	FindEntry(id uuid.UUID) (uuid.UUID, error)
	AddEntry(id uuid.UUID, baseId uuid.UUID) error

	Close()
}
