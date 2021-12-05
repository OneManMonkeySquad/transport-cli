package backends

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

type httpPersistence struct {
	host string
}

func NewHTTP(host string) *httpPersistence {
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	return &httpPersistence{
		host: host,
	}
}

func (p *httpPersistence) Close() {
}

func (p *httpPersistence) UploadFile(fileName string, data []byte) error {
	return errors.New("http backend is read-only")
}

func (p *httpPersistence) DownloadFile(fileName string) ([]byte, error) {
	resp, err := http.Get(p.host + fileName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
