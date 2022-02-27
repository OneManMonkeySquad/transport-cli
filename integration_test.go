package main

import (
	"crypto/sha256"
	"os"
	"testing"

	"github.com/OneManMonkeySquad/transport-cli/backends"
)

func compareDirs(t *testing.T, dir string, dir2 string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	foo := make(map[string][32]byte)
	{
		for _, entry := range entries {
			content, _ := os.ReadFile(entry.Name())
			foo[entry.Name()] = sha256.Sum256(content)
		}
	}

	entries2, err := os.ReadDir(dir2)
	if err != nil {
		t.Fatal(err)
	}

	{
		for _, entry := range entries2 {
			content, _ := os.ReadFile(entry.Name())

			if sha256.Sum256(content) != foo[entry.Name()] {
				t.Errorf("File %v different", entry.Name())
			}
		}
	}
}

func TestBaseRestore(t *testing.T) {
	backend := backends.NewLocal("local_db")
	cfg := NewConfig(backend)
	defer cfg.Backend.Close()

	os.RemoveAll("local_db")
	os.MkdirAll("local_db", 0777)

	os.RemoveAll("out")

	err := version(cfg, "test_data/base1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg.Backend, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = restore(cfg.Backend, "latest", "out")
	if err != nil {
		t.Fatal(err)
	}

	compareDirs(t, "out", "test_data/base1")
}

func TestPatchRestore(t *testing.T) {
	backend := backends.NewLocal("local_db")
	cfg := NewConfig(backend)
	defer cfg.Backend.Close()

	os.RemoveAll("local_db")
	os.MkdirAll("local_db", 0777)

	os.RemoveAll("out")

	err := version(cfg, "test_data/base1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg.Backend, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = patch(cfg, "latest", "test_data/patch1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg.Backend, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = restore(cfg.Backend, "latest", "out")
	if err != nil {
		t.Fatal(err)
	}

	compareDirs(t, "out", "test_data/patch1")
}
