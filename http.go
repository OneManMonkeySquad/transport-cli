package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

type http_persistence struct {
	host string
}

func NewHTTP(host string) *http_persistence {
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	return &http_persistence{
		host: host,
	}
}

func (p *http_persistence) Close() {
}

func (p *http_persistence) UploadFile(fileName string, data []byte) error {
	return errors.New("http backend is read-only")
}

func (p *http_persistence) DownloadFile(fileName string) ([]byte, error) {
	resp, err := http.Get(p.host + fileName)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
