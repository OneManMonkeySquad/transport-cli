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
	switch command {
	case "base":
		base()
	case "patch":
		patch(backend)
	case "commit":
		commit(backend)
	case "restore":
		restore(backend)
	case "tags":
		tags(backend)
	default:
		log.Fatal("Unknown command", command)
	}
}
