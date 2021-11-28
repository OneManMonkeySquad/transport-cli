package main

import (
	"errors"
	"strings"

	"github.com/OneManMonkeySquad/transport-cli/backends"

	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Backend     Backend
	ChunkSizeMb int
}

func NewConfig(backend Backend) *Config {
	return &Config{
		Backend:     backend,
		ChunkSizeMb: 50,
	}
}

func (cfg *Config) ChunkSize() int {
	return cfg.ChunkSizeMb * 1024 * 1024
}

func readConfig() (*Config, error) {
	cfg, err := toml.LoadFile("transport.toml")
	if err != nil {
		return nil, err
	}

	chunkSize := (int)(cfg.Get("chunk_size_mb").(int64))

	backendType := cfg.Get("backend").(string)

	var backend Backend
	if strings.EqualFold(backendType, "sftp") {
		host := cfg.Get("sftp.host").(string)
		user := cfg.GetDefault("sftp.user", "").(string)
		password := cfg.GetDefault("sftp.pw", "").(string)

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.Password(password)},
		}
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

		backend, err = backends.NewSFTP(host, sshConfig)
		if err != nil {
			return nil, err
		}
	} else if strings.EqualFold(backendType, "local") {
		path := cfg.Get("local.path").(string)

		backend = backends.NewLocal(path)
	} else if strings.EqualFold(backendType, "http") {
		host := cfg.Get("http.host").(string)

		backend = backends.NewHTTP(host)
	} else {
		return nil, errors.New("unknown backend '" + backendType + "'")
	}

	config := NewConfig(backend)
	config.ChunkSizeMb = chunkSize
	return config, nil
}
