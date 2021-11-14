package main

import (
	"bytes"
	"io"
	"log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftp_persistence struct {
	client *sftp.Client
}

func NewSFTP(host string, config *ssh.ClientConfig) (*sftp_persistence, error) {
	client, err := connectSFTP(host, config)
	if err != nil {
		return nil, err
	}

	return &sftp_persistence{
		client: client,
	}, nil
}

func (p *sftp_persistence) Close() {
	p.client.Close()
}

func (p *sftp_persistence) UploadFile(fileName string, data []byte) error {
	f, err := p.client.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (p *sftp_persistence) DownloadFile(fileName string) ([]byte, error) {
	f, err := p.client.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func connectSFTP(host string, config *ssh.ClientConfig) (*sftp.Client, error) {
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return client, nil
}
