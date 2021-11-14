package main

import (
	"os"
	"path/filepath"
)

type local_persistence struct {
	path string
}

func NewLocal(path string) *local_persistence {
	return &local_persistence{
		path: path,
	}
}

func (p *local_persistence) Close() {
}

func (p *local_persistence) UploadFile(fileName string, data []byte) error {
	filePath := filepath.Join(p.path, fileName)
	return os.WriteFile(filePath, data, 0644)
}

func (p *local_persistence) DownloadFile(fileName string) ([]byte, error) {
	filePath := filepath.Join(p.path, fileName)
	return os.ReadFile(filePath)
}
