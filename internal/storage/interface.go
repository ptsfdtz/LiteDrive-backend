package storage

import (
	"io"
)

// Storage 存储接口
type Storage interface {
	// Upload 上传文件
	Upload(path string, reader io.Reader, size int64) error

	// Download 下载文件
	Download(path string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(path string) error

	// Exists 检查文件是否存在
	Exists(path string) (bool, error)

	// GetSize 获取文件大小
	GetSize(path string) (int64, error)
}
