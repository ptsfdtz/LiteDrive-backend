package storage

import (
	"io"
	"os"
	"path/filepath"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Upload(path string, reader io.Reader, size int64) error {
	fullPath := filepath.Join(l.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (l *LocalStorage) Download(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, path)
	return os.Open(fullPath)
}

func (l *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(l.basePath, path)
	return os.Remove(fullPath)
}

func (l *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(l.basePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (l *LocalStorage) GetSize(path string) (int64, error) {
	fullPath := filepath.Join(l.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
