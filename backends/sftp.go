package backends

import (
	"bytes"
	"io"
	"log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type sftpPersistence struct {
	client *sftp.Client
}

func NewSFTP(host string, config *ssh.ClientConfig) (*sftpPersistence, error) {
	client, err := connectSFTP(host, config)
	if err != nil {
		return nil, err
	}

	return &sftpPersistence{
		client: client,
	}, nil
}

func (p *sftpPersistence) Close() {
	p.client.Close()
}

func (p *sftpPersistence) UploadFile(fileName string, data []byte) error {
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

func (p *sftpPersistence) DownloadFile(fileName string) ([]byte, error) {
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
