package main

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/OneManMonkeySquad/transport-cli/data_hives"
	"github.com/OneManMonkeySquad/transport-cli/meta_hives"
)

func TestBaseRestore(t *testing.T) {
	metaHive, err := meta_hives.NewSqlite("local_db/test.db")
	if err != nil {
		t.Fatal(err)
	}

	dataHive := data_hives.NewLocal("local_db")
	cfg := NewConfig(metaHive, dataHive)
	defer cfg.dataHive.Close()

	os.RemoveAll("local_db")
	os.MkdirAll("local_db", 0777)

	os.RemoveAll("out")

	err = version(cfg, "test_data/base1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = restore(cfg, "latest", "out")
	if err != nil {
		t.Fatal(err)
	}

	compareDirs(t, "out", "test_data/base1")
}

func TestPatchRestore(t *testing.T) {
	metaHive, err := meta_hives.NewSqlite("local_db/test.db")
	if err != nil {
		t.Fatal(err)
	}

	dataHive := data_hives.NewLocal("local_db")
	cfg := NewConfig(metaHive, dataHive)
	defer cfg.dataHive.Close()

	os.RemoveAll("local_db")
	os.MkdirAll("local_db", 0777)

	os.RemoveAll("out")

	err = version(cfg, "test_data/base1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = patch(cfg, "latest", "test_data/patch1")
	if err != nil {
		t.Fatal(err)
	}

	err = commit(cfg, "latest")
	if err != nil {
		t.Fatal(err)
	}

	err = restore(cfg, "latest", "out")
	if err != nil {
		t.Fatal(err)
	}

	compareDirs(t, "out", "test_data/patch1")
}

func compareDirs(t *testing.T, dir string, dir2 string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	foo := make(map[string][32]byte)
	{
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

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
			if entry.IsDir() {
				compareDirs(t, filepath.Join(dir, entry.Name()), filepath.Join(dir2, entry.Name()))
				continue
			}

			content, _ := os.ReadFile(entry.Name())

			if sha256.Sum256(content) != foo[entry.Name()] {
				t.Errorf("File %v different", entry.Name())
			}
		}
	}
}
