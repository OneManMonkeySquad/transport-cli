package backends

import (
	"os"
	"path/filepath"
)

type localPersistence struct {
	path string
}

func NewLocal(path string) *localPersistence {
	return &localPersistence{
		path: path,
	}
}

func (p *localPersistence) Close() {
}

func (p *localPersistence) UploadFile(fileName string, data []byte) error {
	filePath := filepath.Join(p.path, fileName)
	return os.WriteFile(filePath, data, 0644)
}

func (p *localPersistence) DownloadFile(fileName string) ([]byte, error) {
	filePath := filepath.Join(p.path, fileName)
	return os.ReadFile(filePath)
}
