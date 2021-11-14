package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Backend interface {
	UploadFile(fileName string, data []byte) error
	DownloadFile(fileName string) ([]byte, error)
	Close()
}

// Entries
type BaseEntry struct {
	FileName string
	MD5Hash  string
	Content  []byte
}

type DeletedEntry struct {
	FileName string
}

// Files
type BaseFile struct {
	Version int
	ID      uuid.UUID
	Entries []BaseEntry
}

type PatchFile struct {
	Version int
	ID      uuid.UUID
	BaseID  uuid.UUID
	Changed []BaseEntry
	Deleted []DeletedEntry
}

// Other
type FlatPatch struct {
	Entries []BaseEntry
	Deleted []DeletedEntry
}

func md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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

func writeJson(v interface{}, path string) {
	str, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile(path, str, 0644)
}

func createBase(srcDir string) BaseFile {
	var baseFile BaseFile
	{
		baseFile.Version = 1
		baseFile.ID = uuid.New()

		files, err := os.ReadDir(srcDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			filePath := filepath.Join(srcDir, file.Name())

			hash, err := md5sum(filePath)
			if err != nil {
				log.Fatal(err)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}

			entry := BaseEntry{
				FileName: file.Name(),
				MD5Hash:  hash,
				Content:  content,
			}
			baseFile.Entries = append(baseFile.Entries, entry)
		}
	}
	return baseFile
}

func readBaseFile(path string) BaseFile {
	var baseFile BaseFile
	{
		fileData, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(fileData, &baseFile)
		if err != nil {
			log.Fatal(err)
		}

		if baseFile.Version != 1 {
			log.Fatal("Base file has wrong version")
		}
	}
	return baseFile
}

func createPatch(srcDir string, baseFile BaseFile) (*PatchFile, error) {
	var patch PatchFile
	patch.Version = 1
	patch.ID = uuid.New()
	patch.BaseID = baseFile.ID

	for _, baseEntry := range baseFile.Entries {
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

		hash, err := md5sum(filePath)
		if err != nil {
			return nil, err
		}

		if hash != baseEntry.MD5Hash {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}

			var changed BaseEntry
			changed.FileName = baseEntry.FileName
			changed.MD5Hash = hash
			changed.Content = content
			patch.Changed = append(patch.Changed, changed)
			continue
		}
	}

	if len(patch.Changed) == 0 && len(patch.Deleted) == 0 {
		return nil, errors.New("no changes")
	}

	return &patch, nil
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

type Tag struct {
	Name string
	ID   uuid.UUID
}

type DatabaseEntry struct {
	ID     uuid.UUID
	BaseID uuid.UUID
}

type Database struct {
	Tags    []Tag
	Entries []DatabaseEntry
}

func findTag(db *Database, name string) uuid.UUID {
	for _, tag := range db.Tags {
		if tag.Name == name {
			return tag.ID
		}
	}
	return uuid.Nil
}

func findEntry(db *Database, id uuid.UUID) *DatabaseEntry {
	for _, entry := range db.Entries {
		if entry.ID == id {
			return &entry
		}
	}
	return nil
}

func downloadDatabase(persistence Backend) (*Database, error) {
	dbContent, err := persistence.DownloadFile("db.json")
	if err != nil {
		return nil, err
	}

	var db Database
	err = json.Unmarshal(dbContent, &db)
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func uploadDatabase(db *Database, persistence Backend) error {
	content, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return persistence.UploadFile("db.json", content)
}

func base() {
	if len(os.Args) <= 2 {
		fmt.Println("tp base {directory}")
		return
	}

	base := createBase(os.Args[2])
	writeJson(base, "staging/"+base.ID.String()+".json")
	fmt.Println("base:" + base.ID.String())
}

func patch(persistence Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp patch {tag} {directory}")
		return
	}

	db, err := downloadDatabase(persistence)
	if err == os.ErrNotExist {
		fmt.Println("No Database found. Make sure you have uploaded at least one base patch.")
		return
	} else if err != nil {
		log.Fatal(err)
		return
	}

	tagName := os.Args[2]
	tag := findTag(db, tagName)
	if tag == uuid.Nil {
		log.Fatal("Tag not found", tagName)
		return
	}

	entry := findEntry(db, tag)
	if entry == nil {
		log.Fatal("Entry for tag not found. Existing database not consistent", tagName)
		return
	}

	entryContent, err := persistence.DownloadFile(entry.ID.String() + ".json")
	if err != nil {
		log.Fatal(err)
		return
	}

	var baseFile BaseFile
	err = json.Unmarshal(entryContent, &baseFile)
	if err != nil {
		log.Fatal(err)
	}

	if baseFile.Version != 1 {
		log.Fatal("Base file has wrong version")
		return
	}

	patch, err := createPatch(os.Args[3], baseFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	writeJson(patch, "staging/"+patch.ID.String()+".json")
	fmt.Println("patch:" + patch.ID.String())
}

func commit(persistence Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp commit {tag} {guid}")
		return
	}

	updateTagName := os.Args[2]
	if strings.ContainsAny(updateTagName, " .:;'#+*~") {
		fmt.Println("Tag name contains invalid chars", updateTagName)
		return
	}

	fileTypeAndGuid := os.Args[3]

	var newEntryID uuid.UUID
	var newBaseID uuid.UUID
	var filePath string
	{
		if strings.HasPrefix(fileTypeAndGuid, "base:") {
			filePath = "staging/" + fileTypeAndGuid[5:] + ".json"
			base := readBaseFile(filePath)
			newEntryID = base.ID
			newBaseID = uuid.Nil
		} else if strings.HasPrefix(fileTypeAndGuid, "patch:") {
			filePath = "staging/" + fileTypeAndGuid[6:] + ".json"
			patch := readPatchFile(filePath)
			newEntryID = patch.ID
			newBaseID = patch.BaseID
		} else {
			fmt.Println("Not a valid patch identifier", fileTypeAndGuid)
			return
		}
	}

	//
	db, err := downloadDatabase(persistence)
	if _, ok := err.(*os.PathError); ok {
		fmt.Println("Existing patch database not found, creating one...")
		db = &Database{}
	} else if err == os.ErrNotExist {
		fmt.Println("Existing patch database not found, creating one...")
		db = &Database{}
	} else if err != nil {
		log.Fatal(err)
		return
	}

	// Make sure entry is unique
	for _, entry := range db.Entries {
		if entry.ID == newEntryID {
			log.Fatal("Entry already exists")
			return
		}
	}

	// Update/Insert tag
	foundTag := false
	for i, tag := range db.Tags {
		if tag.Name == updateTagName {
			db.Tags[i].ID = newEntryID
			foundTag = true
			break
		}
	}
	if !foundTag {
		newTag := Tag{
			Name: updateTagName,
			ID:   newEntryID,
		}
		db.Tags = append(db.Tags, newTag)

		fmt.Println("Added new tag", updateTagName)
	}

	// Insert entry
	newEntry := DatabaseEntry{
		ID:     newEntryID,
		BaseID: newBaseID,
	}
	db.Entries = append(db.Entries, newEntry)

	// Upload patch
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = persistence.UploadFile(newEntryID.String()+".json", data)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Upload DB
	err = uploadDatabase(db, persistence)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Remove patch
	os.Remove(filePath)
}

func findRestoreChain(db *Database, head uuid.UUID) []DatabaseEntry {
	var restoreChain []DatabaseEntry
	{
		var currentHead = head
		for currentHead != uuid.Nil {
			entry := findEntry(db, currentHead)
			if entry == nil {
				log.Fatal("Patch not found")
			}

			restoreChain = append(restoreChain, *entry)
			currentHead = entry.BaseID
		}
	}

	for i, j := 0, len(restoreChain)-1; i < j; i, j = i+1, j-1 {
		restoreChain[i], restoreChain[j] = restoreChain[j], restoreChain[i]
	}

	return restoreChain
}

func flattenRestoreChain(db *Database, restoreChain []DatabaseEntry, persistence Backend) (*FlatPatch, error) {
	entryMap := make(map[string]BaseEntry)
	deletedMap := make(map[string]DeletedEntry)

	for i, entry := range restoreChain {
		patchContent, err := persistence.DownloadFile(entry.ID.String() + ".json")
		if err != nil {
			return nil, err
		}

		if i == 0 {
			var baseFile BaseFile
			err = json.Unmarshal(patchContent, &baseFile)
			if err != nil {
				return nil, err
			}

			for _, entry := range baseFile.Entries {
				entryMap[entry.FileName] = entry
			}
		} else {
			var patchFile PatchFile
			err = json.Unmarshal(patchContent, &patchFile)
			if err != nil {
				return nil, err
			}

			for _, entry := range patchFile.Changed {
				entryMap[entry.FileName] = entry
				delete(deletedMap, entry.FileName)
			}
			for _, entry := range patchFile.Deleted {
				delete(entryMap, entry.FileName)
				deletedMap[entry.FileName] = entry
			}
		}
	}

	var result FlatPatch

	result.Entries = make([]BaseEntry, 0, len(entryMap))
	for _, entry := range entryMap {
		result.Entries = append(result.Entries, entry)
	}

	result.Deleted = make([]DeletedEntry, 0, len(deletedMap))
	for _, entry := range deletedMap {
		result.Deleted = append(result.Deleted, entry)
	}

	return &result, nil
}

func restore(persistence Backend) {
	if len(os.Args) <= 3 {
		fmt.Println("tp restore {tag} {dir}")
		return
	}

	db, err := downloadDatabase(persistence)
	if err != nil {
		fmt.Println("Patch database not found.")
		return
	}

	tagName := os.Args[2]
	path := os.Args[3]
	fmt.Println("Restoring '" + tagName + "' to '" + path + "'...")

	head := findTag(db, tagName)

	restoreChain := findRestoreChain(db, head)

	// Now, instead of just going through patches, we collapse them into one.
	// This way we don't write a single file multiple times or write and then delete a file.
	flatPatch, err := flattenRestoreChain(db, restoreChain, persistence)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, entry := range flatPatch.Entries {
		filePath := filepath.Join(path, entry.FileName)

		hash, _ := md5sum(filePath)
		if hash != entry.MD5Hash {
			fmt.Println("New", filePath)
			os.WriteFile(filePath, entry.Content, 0644)
		} else {
			fmt.Println("Equal Hash", filePath)
		}
	}
	for _, entry := range flatPatch.Deleted {
		filePath := filepath.Join(path, entry.FileName)
		fmt.Println("Delete", filePath)

		os.Remove(filePath)
	}
}

func tags(persistence Backend) {
	db, err := downloadDatabase(persistence)
	if err != nil {
		fmt.Println("Patch database not found.")
		return
	}

	for _, tag := range db.Tags {
		fmt.Println(tag.Name)
	}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("tp base {dir}")
		fmt.Println("tp patch {tag} {dir}")
		fmt.Println("tp commit {tag} {patch_guid}")
		fmt.Println("tp restore {tag} {dir}")
		fmt.Println("tp tags")
		return
	}

	backend, err := readConfig()
	if err != nil {
		log.Fatalf("Configuration invalid: %v", err)
		return
	}
	defer backend.Close()

	command := os.Args[1]
	if command == "base" {
		base()
	} else if command == "patch" {
		patch(backend)
	} else if command == "commit" {
		commit(backend)
	} else if command == "restore" {
		restore(backend)
	} else if command == "tags" {
		tags(backend)
	} else {
		fmt.Println("Unknown command", command)
	}
}
