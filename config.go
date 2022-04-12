package main

import (
	"errors"
	"strings"

	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/ssh"

	"github.com/OneManMonkeySquad/transport-cli/data_hives"
	"github.com/OneManMonkeySquad/transport-cli/meta_hives"
)

type Config struct {
	dataHive    DataHive
	metaHive    MetaHive
	chunkSizeMb int
}

func NewConfig(metaHive MetaHive, dataHive DataHive) *Config {
	return &Config{
		dataHive:    dataHive,
		metaHive:    metaHive,
		chunkSizeMb: 50,
	}
}

func (cfg *Config) ChunkSize() int {
	return cfg.chunkSizeMb * 1024 * 1024
}

func readConfig(name string) (*Config, error) {
	cfg, err := toml.LoadFile(name)
	if err != nil {
		return nil, err
	}

	chunkSize := (int)(cfg.Get("chunk_size_mb").(int64))

	dataHiveType := cfg.Get("data_hive").(string)

	var dataHive DataHive
	if strings.EqualFold(dataHiveType, "sftp") {
		host := cfg.Get("sftp.host").(string)
		user := cfg.GetDefault("sftp.user", "").(string)
		password := cfg.GetDefault("sftp.pw", "").(string)
		subfolder := cfg.GetDefault("sftp.subfolder", "").(string)

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.Password(password)},
		}
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

		dataHive, err = data_hives.NewSFTP(host, subfolder, sshConfig)
		if err != nil {
			return nil, err
		}
	} else if strings.EqualFold(dataHiveType, "local") {
		path := cfg.Get("local.path").(string)

		dataHive = data_hives.NewLocal(path)
	} else if strings.EqualFold(dataHiveType, "http") {
		host := cfg.Get("http.host").(string)

		dataHive = data_hives.NewHTTP(host)
	} else if strings.EqualFold(dataHiveType, "s3") {
		// #todo
		dataHive, err = data_hives.NewS3("...", "...", "...", "...", "...")
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unknown data_hive '" + dataHiveType + "'")
	}

	metaHiveType := cfg.Get("meta_hive").(string)

	var metaHive MetaHive
	if strings.EqualFold(metaHiveType, "php") {
		address := cfg.Get("php.address").(string)

		metaHive, err = meta_hives.NewPhp(address)
		if err != nil {
			return nil, err
		}
	} else if strings.EqualFold(metaHiveType, "sqlite") {
		fileName := cfg.Get("sqlite.file_name").(string)

		metaHive, err = meta_hives.NewSqlite(fileName)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unknown meta_hive '" + metaHiveType + "'")
	}

	config := NewConfig(metaHive, dataHive)
	config.chunkSizeMb = chunkSize
	return config, nil
}
