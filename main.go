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
		fmt.Println("tp commit {tag}")
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
		if len(os.Args) <= 2 {
			fmt.Println("tp base {directory}")
			return
		}
		srcDir := os.Args[2]
		err := base(cfg, srcDir)
		if err != nil {
			log.Fatal(err)
		}

	case "patch":
		if len(os.Args) <= 3 {
			fmt.Println("tp patch {tag} {directory}")
			return
		}
		tagName := os.Args[2]
		srcDir := os.Args[3]
		err := patch(cfg, tagName, srcDir)
		if err != nil {
			log.Fatal(err)
		}

	case "commit":
		if len(os.Args) <= 2 {
			fmt.Println("tp commit {tag}")
			return
		}

		tagName := os.Args[2]
		err := commit(cfg.Backend, tagName)
		if err != nil {
			log.Fatal(err)
		}

	case "restore":
		if len(os.Args) <= 3 {
			fmt.Println("tp restore {tag} {dir}")
			return
		}
		tagName := os.Args[2]
		path := os.Args[3]
		err := restore(cfg.Backend, tagName, path)
		if err != nil {
			log.Fatal(err)
		}

	case "tags":
		err := tags(cfg.Backend)
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("Unknown command", command)
	}
}
