package main

import (
	"log"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Version struct {
		Directory string `arg:""`
	} `cmd:"" help:"Create version."`

	Patch struct {
		Tag       string `arg:""`
		Directory string `arg:""`
	} `cmd:"" help:"Create patch relative to latest published version/patch."`

	Commit struct {
		Tag string `arg:""`
	} `cmd:"" help:"Publish staged version/patch."`

	Restore struct {
		Tag       string `arg:""`
		Directory string `arg:""`
	} `cmd:"" help:"Restore the latest published version/release into directory."`

	Tags struct {
	} `cmd:"" help:"Print published tags."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "version <directory>":
		cfg, err := readConfig("production.toml")
		if err != nil {
			log.Fatalf("Configuration invalid: %v", err)
			return
		}
		defer cfg.dataHive.Close()

		err = version(cfg, CLI.Version.Directory)
		if err != nil {
			log.Fatal(err)
		}

	case "patch <tag> <directory>":
		cfg, err := readConfig("production.toml")
		if err != nil {
			log.Fatalf("Configuration invalid: %v", err)
			return
		}
		defer cfg.dataHive.Close()

		err = patch(cfg, CLI.Patch.Tag, CLI.Patch.Directory)
		if err != nil {
			log.Fatal(err)
		}

	case "commit <tag>":
		cfg, err := readConfig("production.toml")
		if err != nil {
			log.Fatalf("Configuration invalid: %v", err)
			return
		}
		defer cfg.dataHive.Close()

		err = commit(cfg, CLI.Commit.Tag)
		if err != nil {
			log.Fatal(err)
		}

	case "restore <tag> <directory>":
		cfg, err := readConfig("release.toml")
		if err != nil {
			log.Fatalf("Configuration invalid: %v", err)
			return
		}
		defer cfg.dataHive.Close()

		err = restore(cfg, CLI.Restore.Tag, CLI.Restore.Directory)
		if err != nil {
			log.Fatal(err)
		}

	case "tags":
		cfg, err := readConfig("release.toml")
		if err != nil {
			log.Fatalf("Configuration invalid: %v", err)
			return
		}
		defer cfg.dataHive.Close()

		err = tags(cfg.metaHive)
		if err != nil {
			log.Fatal(err)
		}

	default:
		panic(ctx.Command())
	}
}
