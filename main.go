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

	cfg, err := readConfig()
	if err != nil {
		log.Fatalf("Configuration invalid: %v", err)
		return
	}
	defer cfg.Backend.Close()

	command := os.Args[1]
	switch command {
	case "base":
		base(cfg)
	case "patch":
		patch(cfg)
	case "commit":
		commit(cfg.Backend)
	case "restore":
		restore(cfg.Backend)
	case "tags":
		tags(cfg.Backend)
	default:
		log.Fatal("Unknown command", command)
	}
}
