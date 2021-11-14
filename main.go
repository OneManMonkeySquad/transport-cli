package main

import (
	"fmt"
	"log"
	"os"
)

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
		log.Fatal("Unknown command", command)
	}
}
