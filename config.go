package main

import (
	"errors"
	"strings"

	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/ssh"
)

func readConfig() (Backend, error) {
	cfg, err := toml.LoadFile("transport.toml")
	if err != nil {
		return nil, err
	}

	backendType := cfg.Get("backend").(string)

	if strings.EqualFold(backendType, "sftp") {
		host := cfg.Get("sftp.host").(string)
		user := cfg.GetDefault("sftp.user", "").(string)
		password := cfg.GetDefault("sftp.pw", "").(string)

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.Password(password)},
		}
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

		backend, err := NewSFTP(host, sshConfig)
		if err != nil {
			return nil, err
		}

		return backend, nil
	} else if strings.EqualFold(backendType, "local") {
		path := cfg.Get("local.path").(string)

		backend := NewLocal(path)

		return backend, nil
	} else if strings.EqualFold(backendType, "http") {
		host := cfg.Get("http.host").(string)

		backend := NewHTTP(host)

		return backend, nil
	} else {
		return nil, errors.New("unknown backend '" + backendType + "'")
	}
}
